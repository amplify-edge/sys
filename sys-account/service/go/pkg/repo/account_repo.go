package repo

import (
	"context"
	"fmt"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
)

func (ad *SysAccountRepo) accountFromClaims(ctx context.Context) (context.Context, *pkg.Account, error) {
	claims, err := ad.ObtainAccessClaimsFromMetadata(ctx, true)
	if err != nil {
		return ctx, nil, err
	}
	ad.log.Debugf("Extracted current user claims: email: %s, role: %v", claims.UserEmail, *claims.Role)
	newCtx := context.WithValue(ctx, sharedAuth.ContextKeyClaims, claims)
	acc, err := ad.getAccountAndRole("", claims.UserEmail)
	if err != nil {
		ad.log.Debugf("Cannot get user's account: %v", err)
		return ctx, nil, status.Errorf(codes.NotFound, "current user not found: %v", err)
	}
	return newCtx, acc, nil
}

func (ad *SysAccountRepo) NewAccount(ctx context.Context, in *pkg.Account) (*pkg.Account, error) {
	if err := ad.allowNewAccount(ctx, in); err != nil {
		return nil, err
	}
	acc, err := ad.store.InsertFromPkgAccountRequest(in)
	if err != nil {
		ad.log.Debugf("error unable to create new account request: %v", err)
		return nil, err
	}
	return ad.getAccountAndRole(acc.ID, "")
}

func (ad *SysAccountRepo) GetAccount(ctx context.Context, in *pkg.GetAccountRequest) (*pkg.Account, error) {
	if in == nil {
		return &pkg.Account{},
			status.Errorf(codes.InvalidArgument, "cannot get user account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	acc, err := ad.allowGetAccount(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (ad *SysAccountRepo) ListAccounts(ctx context.Context, in *pkg.ListAccountsRequest) (*pkg.ListAccountsResponse, error) {
	var limit, cursor int64
	var err error
	if in == nil {
		return &pkg.ListAccountsResponse{}, status.Errorf(codes.InvalidArgument, "cannot list user accounts: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	filter, err := ad.allowListAccount(ctx)
	if err != nil {
		return nil, err
	}
	orderBy := in.OrderBy
	if in.IsDescending {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}
	cursor, err = ad.getCursor(in.CurrentPageId)
	if err != nil {
		return nil, err
	}

	if in.PerPageEntries == 0 {
		limit = dao.DefaultLimit
	}

	accounts, next, err := ad.listAccountsAndRoles(filter, orderBy, limit, cursor)
	if err != nil {
		return nil, err
	}

	return &pkg.ListAccountsResponse{
		Accounts:   accounts,
		NextPageId: fmt.Sprintf("%d", next),
	}, nil
}

// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
func (ad *SysAccountRepo) SearchAccounts(ctx context.Context, in *pkg.SearchAccountsRequest) (*pkg.SearchAccountsResponse, error) {
	var limit, cursor int64
	var err error
	if in == nil {
		return &pkg.SearchAccountsResponse{}, status.Errorf(codes.InvalidArgument, "cannot search user accounts: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	filter, err := ad.allowListAccount(ctx)
	if err != nil {
		return nil, err
	}
	for k, v := range in.Query {
		filter.Params[k] = v
	}
	orderBy := in.SearchParam.OrderBy
	if in.SearchParam.IsDescending {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}
	cursor, err = ad.getCursor(in.SearchParam.CurrentPageId)
	if err != nil {
		return nil, err
	}
	if in.SearchParam.PerPageEntries == 0 {
		limit = dao.DefaultLimit
	}
	accounts, next, err := ad.listAccountsAndRoles(filter, orderBy, limit, cursor)
	if err != nil {
		return nil, err
	}
	return &pkg.SearchAccountsResponse{
		SearchResponse: &pkg.ListAccountsResponse{
			Accounts:   accounts,
			NextPageId: fmt.Sprintf("%d", next),
		},
	}, nil
}

func (ad *SysAccountRepo) AssignAccountToRole(ctx context.Context, in *pkg.AssignAccountToRoleRequest) (*pkg.Account, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot assign user Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	err := ad.allowAssignToRole(ctx, in)
	if err != nil {
		return nil, err
	}
	req, err := ad.store.FromPkgRole(&in.Role, in.AssignedAccountId)
	if err != nil {
		return nil, err
	}
	err = ad.store.UpdateRole(req)
	if err != nil {
		return nil, err
	}
	return ad.getAccountAndRole(in.AssignedAccountId, "")
}

func (ad *SysAccountRepo) UpdateAccount(ctx context.Context, in *pkg.Account) (*pkg.Account, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot update Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	cur, err := ad.allowGetAccount(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	ad.log.Debugf("current to be updated user: %v", cur)
	acc, err := ad.store.GetAccount(&coresvc.QueryParams{Params: map[string]interface{}{"id": cur.Id}})
	if err != nil {
		return nil, err
	}
	if in.Disabled {
		acc.Disabled = true
	}
	if !in.Disabled {
		acc.Disabled = false
	}
	if in.Fields != nil && in.Fields.Fields != nil {
		cur.Fields = in.Fields
	}
	if in.Survey != nil && in.Survey.Fields != nil {
		cur.Survey = in.Survey
	}
	if in.Verified {
		acc.Verified = true
	}
	acc.UpdatedAt = timestampNow()
	err = ad.store.UpdateAccount(acc)
	if err != nil {
		ad.log.Debugf("unable to update account: %v", err)
		return nil, err
	}
	return ad.getAccountAndRole(in.Id, "")
}

func (ad *SysAccountRepo) DisableAccount(ctx context.Context, in *pkg.DisableAccountRequest) (*pkg.Account, error) {
	if in == nil {
		return &pkg.Account{}, status.Errorf(codes.InvalidArgument, "cannot update Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	acc, err := ad.allowGetAccount(ctx, in.AccountId)
	if err != nil {
		return nil, err
	}
	acc.Disabled = true
	req, err := ad.store.FromPkgAccount(acc)
	if err != nil {
		return nil, err
	}
	err = ad.store.UpdateAccount(req)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
