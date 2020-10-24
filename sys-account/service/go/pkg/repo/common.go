package repo

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

func (ad *SysAccountRepo) getAccountAndRole(id string) (*pkg.Account, error) {
	// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
	acc, err := ad.store.GetAccount(&coredb.QueryParams{Params: map[string]interface{}{
		"id": id,
	}})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find user account: %v", sharedAuth.Error{Reason: sharedAuth.ErrAccountNotFound})
	}
	role, err := ad.store.GetRole(&coredb.QueryParams{Params: map[string]interface{}{
		"id": acc.RoleId,
	}})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find user role: %v", sharedAuth.Error{Reason: sharedAuth.ErrAccountNotFound})
	}
	userRole, err := role.ToPkgRole()
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find user role: %v", sharedAuth.Error{Reason: sharedAuth.ErrAccountNotFound})
	}

	return acc.ToPkgAccount(userRole)
}

func (ad *SysAccountRepo) listAccountsAndRoles(filter *coredb.QueryParams, orderBy string, limit, cursor int64) ([]*pkg.Account, *int64, error) {
	listAccounts, next, err := ad.store.ListAccount(filter, orderBy, limit, cursor)
	if err != nil {
		return nil, nil, err
	}
	var accounts []*pkg.Account

	for _, acc := range listAccounts {
		r, err := ad.store.GetRole(&coredb.QueryParams{Params: map[string]interface{}{"account_id": acc.ID}})
		if err != nil {
			return nil, nil, err
		}
		role, err := r.ToPkgRole()
		if err != nil {
			return nil, nil, err
		}
		account, err := acc.ToPkgAccount(role)
		if err != nil {
			return nil, nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, &next, nil
}

func (ad *SysAccountRepo) getCursor(currentCursor string) (int64, error) {
	if currentCursor != "" {
		return strconv.ParseInt(currentCursor, 10, 64)
	} else {
		return 0, nil
	}
}
