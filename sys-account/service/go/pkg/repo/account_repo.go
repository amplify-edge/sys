package repo

import (
	"context"
	"fmt"
	"github.com/VictoriaMetrics/metrics"
	"github.com/amplify-cms/sys/sys-account/service/go/pkg/telemetry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	utilities "github.com/amplify-cms/sys-share/sys-core/service/config"
	coresvc "github.com/amplify-cms/sys/sys-core/service/go/pkg/coredb"

	"github.com/amplify-cms/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/amplify-cms/sys-share/sys-account/service/go/pkg/shared"

	"github.com/amplify-cms/sys/sys-account/service/go/pkg/dao"
)

func (ad *SysAccountRepo) accountFromClaims(ctx context.Context) (context.Context, *pkg.Account, error) {
	claims, err := ad.ObtainAccessClaimsFromMetadata(ctx, true)
	if err != nil {
		return ctx, nil, err
	}
	ad.log.Debugf("Extracted current user claims: email: %s, role: %v", claims.UserEmail, claims.Role)
	newCtx := context.WithValue(ctx, sharedAuth.ContextKeyClaims, claims)
	acc, err := ad.getAccountAndRole("", claims.UserEmail)
	if err != nil {
		return ctx, nil, status.Errorf(codes.NotFound, "current user not found: %v", err)
	}
	return newCtx, acc, nil
}

func (ad *SysAccountRepo) NewAccount(ctx context.Context, in *pkg.AccountNewRequest) (*pkg.Account, error) {
	if err := ad.allowNewAccount(ctx, in); err != nil {
		ad.log.Debugf("creation of new account failed: %v", err)
		return nil, err
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
	acc, err := ad.store.InsertFromPkgAccountRequest(in, false)
	if err != nil {
		ad.log.Debugf("error unable to create new account request: %v", err)
		return nil, err
	}
	return ad.getAccountAndRole(acc.ID, "")
}

func (ad *SysAccountRepo) GetAccount(ctx context.Context, in *pkg.IdRequest) (*pkg.Account, error) {
	if in == nil {
		return &pkg.Account{},
			status.Errorf(codes.InvalidArgument, "cannot get user account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	acc, err := ad.allowGetAccount(ctx, in)
	if err != nil {
		return nil, err
	}
	avatar, err := ad.frepo.DownloadFile("", acc.AvatarResourceId)
	if err != nil {
		return nil, err
	}
	acc.Avatar = avatar.Binary
	return acc, nil
}

func (ad *SysAccountRepo) ListAccounts(ctx context.Context, in *pkg.ListAccountsRequest) (*pkg.ListAccountsResponse, error) {
	var limit, cursor int64
	var err error
	if in == nil {
		return &pkg.ListAccountsResponse{}, status.Errorf(codes.InvalidArgument, "cannot list user accounts: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
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

	accounts, next, err := ad.listAccountsAndRoles(filter, orderBy, limit, cursor, in.Matcher)
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
	accounts, next, err := ad.listAccountsAndRoles(filter, orderBy, limit, cursor, in.SearchParam.Matcher)
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
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	roles, err := ad.store.FetchRoles(in.AssignedAccountId)
	if err != nil {
		return nil, err
	}
	// ORG ADMIN
	if in.Role.OrgID != "" && in.Role.ProjectID == "" {
		for _, r := range roles {
			// can only assign for their own org
			if r.OrgId == in.Role.OrgID {
				if err := ad.store.UpdateRole(&dao.Role{
					ID:        r.ID,
					AccountId: r.AccountId,
					Role:      int(in.Role.Role),
					ProjectId: r.ProjectId,
					OrgId:     in.Role.OrgID,
					CreatedAt: r.CreatedAt,
					UpdatedAt: utilities.CurrentTimestamp(),
				}); err != nil {
					return nil, err
				}
				return ad.getAccountAndRole(in.AssignedAccountId, "")
			}
		}
	}
	// PROJECT ADMIN
	if (in.Role.OrgID != "" && in.Role.ProjectID != "") || (in.Role.OrgID == "" && in.Role.ProjectID != "") {
		for _, r := range roles {
			// can only assign for their own project
			if r.OrgId == in.Role.OrgID && r.ProjectId == in.Role.ProjectID {
				if err := ad.store.UpdateRole(&dao.Role{
					ID:        r.ID,
					AccountId: r.AccountId,
					Role:      int(in.Role.Role),
					ProjectId: in.Role.ProjectID,
					OrgId:     in.Role.OrgID,
					CreatedAt: r.CreatedAt,
					UpdatedAt: utilities.CurrentTimestamp(),
				}); err != nil {
					return nil, err
				}
				return ad.getAccountAndRole(in.AssignedAccountId, "")
			}
		}
	}
	// SUPERADMIN
	if in.Role.Role == pkg.SUPERADMIN && sharedAuth.IsSuperadmin(curAcc.Role) {
		for _, r := range roles {
			if err = ad.store.DeleteRole(r.ID); err != nil {
				return nil, err
			}
		}
		newRole := &dao.Role{
			ID:        utilities.NewID(),
			AccountId: in.AssignedAccountId,
			Role:      int(in.Role.Role),
			ProjectId: "",
			OrgId:     "",
			CreatedAt: utilities.CurrentTimestamp(),
			UpdatedAt: utilities.CurrentTimestamp(),
		}
		if err = ad.store.InsertRole(newRole); err != nil {
			return nil, err
		}
		return ad.getAccountAndRole(in.AssignedAccountId, "")
	} else if sharedAuth.IsSuperadmin(curAcc.Role) {
		if len(roles) == 1 && roles[0].Role == int(pkg.GUEST) {
			ad.log.Debug("deleting user guest role")
			if err := ad.store.DeleteRole(roles[0].ID); err != nil {
				return nil, err
			}
		}
		for _, r := range roles {
			// if exists, update current record
			if r.OrgId == in.Role.OrgID && r.ProjectId == in.Role.ProjectID {
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
				return ad.getAccountAndRole(in.AssignedAccountId, "")
			}
		}
		newRole := &dao.Role{
			ID:        utilities.NewID(),
			AccountId: in.AssignedAccountId,
			Role:      int(in.Role.Role),
			ProjectId: in.Role.ProjectID,
			OrgId:     in.Role.OrgID,
			CreatedAt: utilities.CurrentTimestamp(),
			UpdatedAt: utilities.CurrentTimestamp(),
		}
		if err := ad.store.InsertRole(newRole); err != nil {
			return nil, err
		}
		go func() {
			joinedMetrics := metrics.GetOrCreateCounter(fmt.Sprintf(telemetry.JoinProjectLabel, telemetry.METRICS_JOINED_PROJECT, in.Role.OrgID, in.Role.ProjectID))
			joinedMetrics.Inc()
		}()

		return ad.getAccountAndRole(in.AssignedAccountId, "")
	}

	// Regular Users can only allow themselves to be regular user in other project
	if sharedAuth.AllowSelf(curAcc, in.AssignedAccountId) {
		requestedRole := in.Role.Role
		if requestedRole != pkg.USER {
			return nil, status.Errorf(codes.InvalidArgument, "cannot update role: invalid role is specified")
		}
		for _, r := range roles {
			if r.OrgId == in.Role.OrgID && r.ProjectId == in.Role.ProjectID {
				return ad.getAccountAndRole(in.AssignedAccountId, "")
			}
		}
		if err = ad.store.InsertRole(&dao.Role{
			ID:        utilities.NewID(),
			AccountId: in.AssignedAccountId,
			Role:      int(requestedRole),
			ProjectId: in.Role.ProjectID,
			OrgId:     in.Role.OrgID,
			CreatedAt: utilities.CurrentTimestamp(),
			UpdatedAt: utilities.CurrentTimestamp(),
		}); err != nil {
			return nil, status.Errorf(codes.Internal, "cannot append role: %v", err)
		}

		go func() {
			joinedMetrics := metrics.GetOrCreateCounter(fmt.Sprintf(telemetry.JoinProjectLabel, telemetry.METRICS_JOINED_PROJECT, in.Role.OrgID, in.Role.ProjectID))
			joinedMetrics.Inc()
		}()

	}

	return nil, status.Errorf(codes.InvalidArgument, "cannot update role: invalid role is specified")
}

func (ad *SysAccountRepo) UpdateAccount(ctx context.Context, in *pkg.AccountUpdateRequest) (*pkg.Account, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot update Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	cur, err := ad.allowGetAccount(ctx, &pkg.IdRequest{Id: in.Id})
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
	return ad.getAccountAndRole(in.Id, "")
}

func (ad *SysAccountRepo) DisableAccount(ctx context.Context, in *pkg.DisableAccountRequest) (*pkg.Account, error) {
	if in == nil {
		return &pkg.Account{}, status.Errorf(codes.InvalidArgument, "cannot update Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	acc, err := ad.allowGetAccount(ctx, &pkg.IdRequest{Id: in.AccountId})
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

func (ad *SysAccountRepo) DeleteAccount(ctx context.Context, in *pkg.DisableAccountRequest) (*emptypb.Empty, error) {
	if in == nil {
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "cannot update Account: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	_, err := ad.allowGetAccount(ctx, &pkg.IdRequest{Id: in.AccountId})
	if err != nil {
		return nil, err
	}
	err = ad.store.DeleteAccount(in.AccountId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
