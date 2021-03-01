package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/VictoriaMetrics/metrics"
	"go.amplifyedge.org/sys-v2/sys-account/service/go/pkg/telemetry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	utilities "go.amplifyedge.org/sys-share-v2/sys-core/service/config"
	coresvc "go.amplifyedge.org/sys-v2/sys-core/service/go/pkg/coredb"

	sharedAuth "go.amplifyedge.org/sys-share-v2/sys-account/service/go/pkg/shared"
	rpc "go.amplifyedge.org/sys-share-v2/sys-account/service/go/rpc/v2"

	"go.amplifyedge.org/sys-v2/sys-account/service/go/pkg/dao"
)

func (ad *SysAccountRepo) accountFromClaims(ctx context.Context) (context.Context, *rpc.Account, error) {
	claims, err := ad.ObtainAccessClaimsFromMetadata(ctx, true)
	if err != nil {
		return ctx, nil, err
	}
	ad.log.Debugf("Extracted current user claims: email: %s, role: %v", claims.UserEmail, claims.Role)
	newCtx := context.WithValue(ctx, sharedAuth.ContextKeyClaims, claims)
	acc, err := ad.getAccountAndRole(newCtx, "", claims.UserEmail)
	if err != nil {
		return ctx, nil, status.Errorf(codes.NotFound, "current user not found: %v", err)
	}
	return newCtx, acc, nil
}

func (ad *SysAccountRepo) NewAccount(ctx context.Context, in *rpc.AccountNewRequest) (*rpc.Account, error) {
	if err := ad.allowNewAccount(ctx, in); err != nil {
		ad.log.Debugf("creation of new account failed: %v", err)
		return nil, err
	}
	if sharedAuth.IsSuperadmin(in.Roles) {
		ad.log.Debugf("user wanted to create superadmin, fail here as it's not allowed")
		return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{
			Reason: sharedAuth.ErrInvalidParameters,
			Err:    errors.New("superuser creation not allowed"),
		}.Error())
	}
	var logoBytes []byte
	var err error
	if in.AvatarUploadBytes != "" {
		logoBytes, err = utilities.DecodeB64(in.AvatarUploadBytes)
	}
	fresp, err := ad.frepo.UploadFile(in.AvatarFilepath, logoBytes)
	if err != nil {
		return nil, err
	}
	ad.log.Debugf("Uploaded File from path: %s, id: %s", in.AvatarFilepath, fresp.GetId())
	// Remember this is the key to success
	in.AvatarFilepath = fresp.ResourceId
	acc, err := ad.store.InsertFromRpcAccountRequest(in, false)
	if err != nil {
		ad.log.Debugf("error unable to create new account request: %v", err)
		return nil, err
	}
	return ad.getAccountAndRole(ctx, acc.ID, "")
}

func (ad *SysAccountRepo) GetAccount(ctx context.Context, in *rpc.IdRequest) (*rpc.Account, error) {
	if in == nil {
		return &rpc.Account{},
			status.Errorf(codes.InvalidArgument, "cannot get user account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	acc, err := ad.allowGetAccount(ctx, in)
	if err != nil {
		return nil, err
	}
	if acc.AvatarResourceId != "" {
		avatar, err := ad.frepo.DownloadFile("", acc.AvatarResourceId)
		if err != nil {
			return nil, err
		}
		acc.Avatar = avatar.Binary
	}
	return acc, nil
}

func (ad *SysAccountRepo) ListAccounts(ctx context.Context, in *rpc.ListAccountsRequest) (*rpc.ListAccountsResponse, error) {
	var limit, cursor int64
	var err error
	if in == nil {
		return &rpc.ListAccountsResponse{}, status.Errorf(codes.InvalidArgument, "cannot list user accounts: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	limit = in.PerPageEntries
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

	accounts, next, err := ad.listAccountsAndRoles(ctx, filter, orderBy, limit, cursor, in.Matcher)
	if err != nil {
		return nil, err
	}

	return &rpc.ListAccountsResponse{
		Accounts:   accounts,
		NextPageId: fmt.Sprintf("%d", next),
	}, nil
}

// TODO @gutterbacon: In the absence of actual enforcement policy function, this method is a stub. We allow everyone to query anything at this point.
func (ad *SysAccountRepo) SearchAccounts(ctx context.Context, in *rpc.SearchAccountsRequest) (*rpc.SearchAccountsResponse, error) {
	var limit, cursor int64
	var err error
	if in == nil {
		return &rpc.SearchAccountsResponse{}, status.Errorf(codes.InvalidArgument, "cannot search user accounts: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	filter, err := ad.allowListAccount(ctx)
	if err != nil {
		return nil, err
	}
	query := map[string]interface{}{}
	if err := utilities.UnmarshalJson(in.GetQuery(), &query); err != nil {
		return nil, err
	}
	for k, v := range query {
		filter.Params[k] = v
	}
	orderBy := in.GetSearchParams().GetOrderBy()
	if in.GetSearchParams().IsDescending {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}
	cursor, err = ad.getCursor(in.GetSearchParams().CurrentPageId)
	if err != nil {
		return nil, err
	}
	if in.GetSearchParams().PerPageEntries == 0 {
		limit = dao.DefaultLimit
	}
	accounts, next, err := ad.listAccountsAndRoles(ctx, filter, orderBy, limit, cursor, in.GetSearchParams().Matcher)
	if err != nil {
		return nil, err
	}
	return &rpc.SearchAccountsResponse{
		SearchResponse: &rpc.ListAccountsResponse{
			Accounts:   accounts,
			NextPageId: fmt.Sprintf("%d", next),
		},
	}, nil
}

func (ad *SysAccountRepo) AssignAccountToRole(ctx context.Context, in *rpc.AssignAccountToRoleRequest) (*rpc.Account, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot assign user Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	err := ad.allowAssignToRole(ctx, in)
	if err != nil {
		return nil, err
	}
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	roles, err := ad.store.FetchRoles(in.AssignedAccountId)
	if err != nil {
		return nil, err
	}
	// ORG ADMIN
	if in.Role.OrgId != "" && in.Role.ProjectId == "" {
		for _, r := range roles {
			// can only assign for their own org
			if r.OrgId == in.Role.OrgId {
				if err := ad.store.UpdateRole(&dao.Role{
					ID:        r.ID,
					AccountId: r.AccountId,
					Role:      int(in.Role.Role),
					ProjectId: r.ProjectId,
					OrgId:     in.Role.OrgId,
					CreatedAt: r.CreatedAt,
					UpdatedAt: utilities.CurrentTimestamp(),
				}); err != nil {
					return nil, err
				}
				return ad.getAccountAndRole(ctx, in.AssignedAccountId, "")
			}
		}
	}
	// PROJECT ADMIN
	if (in.Role.OrgId != "" && in.Role.ProjectId != "") || (in.Role.OrgId == "" && in.Role.ProjectId != "") {
		for _, r := range roles {
			// can only assign for their own project
			if r.OrgId == in.Role.OrgId && r.ProjectId == in.Role.ProjectId {
				if err := ad.store.UpdateRole(&dao.Role{
					ID:        r.ID,
					AccountId: r.AccountId,
					Role:      int(in.Role.Role),
					ProjectId: in.Role.ProjectId,
					OrgId:     in.Role.OrgId,
					CreatedAt: r.CreatedAt,
					UpdatedAt: utilities.CurrentTimestamp(),
				}); err != nil {
					return nil, err
				}
				return ad.getAccountAndRole(ctx, in.AssignedAccountId, "")
			}
		}
	}
	// SUPERADMIN
	if in.Role.Role == rpc.Roles_SUPERADMIN && sharedAuth.IsSuperadmin(curAcc.GetRoles()) {
		return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{
			Reason: sharedAuth.ErrInvalidParameters,
			Err:    errors.New("superadmin is not assignable"),
		}.Error())
	} else if sharedAuth.IsSuperadmin(curAcc.GetRoles()) {
		if len(roles) == 1 && roles[0].Role == int(rpc.Roles_GUEST) {
			ad.log.Debug("deleting user guest role")
			if err := ad.store.DeleteRole(roles[0].ID); err != nil {
				return nil, err
			}
		}
		for _, r := range roles {
			// if exists, update current record
			if r.OrgId == in.Role.OrgId && r.ProjectId == in.Role.ProjectId {
				if err = ad.store.UpdateRole(&dao.Role{
					ID:        r.ID,
					AccountId: r.AccountId,
					Role:      int(in.Role.Role),
					ProjectId: r.ProjectId,
					OrgId:     r.OrgId,
					CreatedAt: r.CreatedAt,
					UpdatedAt: utilities.CurrentTimestamp(),
				}); err != nil {
					return nil, err
				}
				return ad.getAccountAndRole(ctx, in.AssignedAccountId, "")
			}
		}
		newRole := &dao.Role{
			ID:        utilities.NewID(),
			AccountId: in.AssignedAccountId,
			Role:      int(in.Role.Role),
			ProjectId: in.Role.ProjectId,
			OrgId:     in.Role.OrgId,
			CreatedAt: utilities.CurrentTimestamp(),
			UpdatedAt: utilities.CurrentTimestamp(),
		}
		if err := ad.store.InsertRole(newRole); err != nil {
			return nil, err
		}
		go func() {
			joinedMetrics := metrics.GetOrCreateCounter(fmt.Sprintf(telemetry.JoinProjectLabel, telemetry.METRICS_JOINED_PROJECT, in.Role.OrgId, in.Role.ProjectId))
			joinedMetrics.Inc()
		}()

		return ad.getAccountAndRole(ctx, in.AssignedAccountId, "")
	}

	// Regular Users can only allow themselves to be regular user in other project
	if sharedAuth.AllowSelf(curAcc, in.AssignedAccountId) {
		requestedRole := in.Role.Role
		if requestedRole != rpc.Roles_USER {
			return nil, status.Errorf(codes.InvalidArgument, "cannot update role: invalid role is specified")
		}
		for _, r := range roles {
			if r.OrgId == in.Role.OrgId && r.ProjectId == in.Role.ProjectId {
				return ad.getAccountAndRole(ctx, in.AssignedAccountId, "")
			}
		}
		if err = ad.store.InsertRole(&dao.Role{
			ID:        utilities.NewID(),
			AccountId: in.AssignedAccountId,
			Role:      int(requestedRole),
			ProjectId: in.Role.ProjectId,
			OrgId:     in.Role.OrgId,
			CreatedAt: utilities.CurrentTimestamp(),
			UpdatedAt: utilities.CurrentTimestamp(),
		}); err != nil {
			return nil, status.Errorf(codes.Internal, "cannot append role: %v", err)
		}

		go func() {
			joinedMetrics := metrics.GetOrCreateCounter(fmt.Sprintf(telemetry.JoinProjectLabel, telemetry.METRICS_JOINED_PROJECT, in.Role.OrgId, in.Role.ProjectId))
			joinedMetrics.Inc()
		}()

	}

	return nil, status.Errorf(codes.InvalidArgument, "cannot update role: invalid role is specified")
}

func (ad *SysAccountRepo) UpdateAccount(ctx context.Context, in *rpc.AccountUpdateRequest) (*rpc.Account, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot update Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	cur, err := ad.allowGetAccount(ctx, &rpc.IdRequest{Id: in.Id})
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
	if in.Verified {
		acc.Verified = true
	}
	if in.AvatarFilepath != "" && len(in.AvatarUploadBytes) != 0 {
		updatedAvatar, err := ad.frepo.UploadFile(in.AvatarFilepath, in.AvatarUploadBytes)
		if err != nil {
			return nil, err
		}
		acc.AvatarResourceId = updatedAvatar.ResourceId
	}

	acc.UpdatedAt = utilities.CurrentTimestamp()
	err = ad.store.UpdateAccount(acc)
	if err != nil {
		ad.log.Debugf("unable to update account: %v", err)
		return nil, err
	}
	return ad.getAccountAndRole(ctx, in.Id, "")
}

func (ad *SysAccountRepo) DisableAccount(ctx context.Context, in *rpc.DisableAccountRequest) (*rpc.Account, error) {
	if in == nil {
		return &rpc.Account{}, status.Errorf(codes.InvalidArgument, "cannot update Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	acc, err := ad.allowGetAccount(ctx, &rpc.IdRequest{Id: in.AccountId})
	if err != nil {
		return nil, err
	}
	acc.Disabled = true
	req, err := ad.store.FromRpcAccount(acc)
	if err != nil {
		return nil, err
	}
	err = ad.store.UpdateAccount(req)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (ad *SysAccountRepo) DeleteAccount(ctx context.Context, in *rpc.DisableAccountRequest) (*emptypb.Empty, error) {
	if in == nil {
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "cannot update Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	_, err := ad.allowGetAccount(ctx, &rpc.IdRequest{Id: in.AccountId})
	if err != nil {
		return nil, err
	}
	err = ad.store.DeleteAccount(in.AccountId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
