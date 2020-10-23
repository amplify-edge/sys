package repo

import (
	"context"
	"fmt"
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
	return ad.getAccountAndRole(in.Id)
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
		return &pkg.SearchAccountsResponse{}, status.Errorf(codes.InvalidArgument, "cannot search user accounts: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}
	filter := &coredb.QueryParams{Params: map[string]interface{}{}}
	for k, v := range in.Query {
		filter.Params[k] = v
	}
	orderBy := in.SearchParam.OrderBy + " ASC"
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

// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
func (ad *SysAccountRepo) AssignAccountToRole(ctx context.Context, in *pkg.AssignAccountToRoleRequest) (*pkg.Account, error) {
	if in == nil {
		return &pkg.Account{}, status.Errorf(codes.InvalidArgument, "cannot assign user Account: %v", auth.Error{Reason: auth.ErrInvalidParameters})
	}
	// acc, err := ad.getAccountAndRole(in.AssignedAccountId)
	// if err != nil {
	// 	return nil, err
	// }

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
