package repo

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

func (ad *SysAccountRepo) accountFromClaims(ctx context.Context) (context.Context, *pkg.Account, error) {
	claims, err := ad.ObtainAccessClaimsFromMetadata(ctx, true)
	if err != nil {
		return ctx, nil, err
	}
	ad.log.Debugf("Extracted claims: user_id: %s, email: %s, role: %v", claims.Id, claims.UserEmail, *claims.Role)
	newCtx := context.WithValue(ctx, sharedAuth.ContextKeyClaims, claims)
	acc, err := ad.getAccountAndRole("", claims.UserEmail)
	if err != nil {
		return ctx, nil, status.Errorf(codes.NotFound, "current user not found: %v", err)
	}
	return newCtx, acc, nil
}

func (ad *SysAccountRepo) NewAccount(ctx context.Context, in *pkg.Account) (*pkg.Account, error) {
	if err := ad.allowNewAccount(ctx, in); err != nil {
		return nil, err
	}
	now := timestampNow()
	roleId := coredb.NewID()
	if err := ad.store.InsertRole(&dao.Role{
		ID:        roleId,
		AccountId: in.Id,
		Role:      int(in.Role.Role),
		ProjectId: in.Role.ProjectID,
		OrgId:     in.Role.OrgID,
		CreatedAt: now,
	}); err != nil {
		return nil, err
	}
	if err := ad.store.InsertAccount(&dao.Account{
		ID:                in.Id,
		Email:             in.Email,
		Password:          in.Password,
		RoleId:            roleId,
		UserDefinedFields: in.Fields.Fields,
		CreatedAt:         in.CreatedAt,
		UpdatedAt:         in.UpdatedAt,
		LastLogin:         in.LastLogin,
		Disabled:          in.Disabled,
	}); err != nil {
		return nil, err
	}
	acc, err := ad.store.GetAccount(&coredb.QueryParams{Params: map[string]interface{}{"id": in.Id}})
	if err != nil {
		return nil, err
	}
	role, err := ad.store.GetRole(&coredb.QueryParams{Params: map[string]interface{}{"id": acc.RoleId}})
	if err != nil {
		return nil, err
	}
	userRole, err := role.ToPkgRole()
	if err != nil {
		return nil, err
	}
	return acc.ToPkgAccount(userRole)
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
	if cur.Role != in.Role {
		req, err := ad.store.FromPkgRole(in.Role, in.Id)
		if err != nil {
			return nil, err
		}
		err = ad.store.UpdateRole(req)
	}
	req, err := ad.store.FromPkgAccount(in)
	if err != nil {
		return nil, err
	}
	req.UpdatedAt = timestampNow()
	err = ad.store.UpdateAccount(req)
	if err != nil {
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
