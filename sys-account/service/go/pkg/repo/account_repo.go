package repo

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/auth"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

func (ad *SysAccountRepo) NewAccount(ctx context.Context, in *pkg.Account) (*pkg.Account, error) {
	now := timestampNow()
	roleId := coredb.NewID()
	if err := ad.store.InsertRole(&dao.Role{
		ID:        roleId,
		AccountId: in.Id, // TODO @gutterbacon check for uniqueness
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
			status.Errorf(codes.InvalidArgument, "cannot get user account: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}

	// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
	acc, err := ad.store.GetAccount(&coredb.QueryParams{Params: map[string]interface{}{
		"id": in.Id,
	}})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find user account: %v", auth.Error{Reason: auth.ErrAccountNotFound})
	}
	role, err := ad.store.GetRole(&coredb.QueryParams{Params: map[string]interface{}{
		"id": acc.RoleId,
	}})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find user role: %v", auth.Error{Reason: auth.ErrAccountNotFound})
	}
	userRole, err := role.ToPkgRole()
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find user role: %v", auth.Error{Reason: auth.ErrAccountNotFound})
	}

	return acc.ToPkgAccount(userRole)
}

// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
func (ad *SysAccountRepo) ListAccounts(ctx context.Context, in *pkg.ListAccountsRequest) (*pkg.ListAccountsResponse, error) {
	var limit, cursor int64
	var err error
	if in == nil {
		return &pkg.ListAccountsResponse{}, status.Errorf(codes.InvalidArgument, "cannot list user accounts: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}
	filter := &coredb.QueryParams{Params: map[string]interface{}{}}
	orderBy := in.OrderBy + " ASC"
	if in.CurrentPageId != "" {
		cursor, err = strconv.ParseInt(in.CurrentPageId, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "cannot list user accounts: %v", err)
		}
	} else {
		cursor = 0
	}
	if in.PerPageEntries == 0 {
		limit = dao.DefaultLimit
	}
	listAccounts, next, err := ad.store.ListAccount(filter, orderBy, limit, cursor)
	if err != nil {
		return nil, err
	}
	var accounts []*pkg.Account

	for _, acc := range listAccounts {
		r, err := ad.store.GetRole(&coredb.QueryParams{Params: map[string]interface{}{"account_id": acc.ID}})
		if err != nil {
			return nil, err
		}
		role, err := r.ToPkgRole()
		if err != nil {
			return nil, err
		}
		account, err := acc.ToPkgAccount(role)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
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
		return &pkg.SearchAccountsResponse{}, status.Errorf(codes.InvalidArgument, "cannot search user accounts: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}
	filter := &coredb.QueryParams{Params: map[string]interface{}{}}
	for k, v := range in.Query {
		filter.Params[k] = v
	}
	orderBy := in.SearchParam.OrderBy + " ASC"
	if in.SearchParam.CurrentPageId != "" {
		cursor, err = strconv.ParseInt(in.SearchParam.CurrentPageId, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "cannot list user accounts: %v", err)
		}
	} else {
		cursor = 0
	}
	if in.SearchParam.PerPageEntries == 0 {
		limit = dao.DefaultLimit
	}
	listAccounts, next, err := ad.store.ListAccount(filter, orderBy, limit, cursor)
	if err != nil {
		return nil, err
	}
	var accounts []*pkg.Account

	for _, acc := range listAccounts {
		r, err := ad.store.GetRole(&coredb.QueryParams{Params: map[string]interface{}{"account_id": acc.ID}})
		if err != nil {
			return nil, err
		}
		role, err := r.ToPkgRole()
		if err != nil {
			return nil, err
		}
		account, err := acc.ToPkgAccount(role)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return &pkg.SearchAccountsResponse{
		SearchResponse: &pkg.ListAccountsResponse{
			Accounts:   accounts,
			NextPageId: fmt.Sprintf("%d", next),
		},
	}, nil
}

// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
func (ad *SysAccountRepo) AssignAccountToRole(context.Context, *pkg.AssignAccountToRoleRequest) (*pkg.Account, error) {
	return &pkg.Account{}, nil
}

// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
func (ad *SysAccountRepo) UpdateAccount(context.Context, *pkg.Account) (*pkg.Account, error) {
	return &pkg.Account{}, nil
}

// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
func (ad *SysAccountRepo) DisableAccount(context.Context, *pkg.DisableAccountRequest) (*pkg.Account, error) {
	return &pkg.Account{}, nil
}
