package repo

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	coresvc "go.amplifyedge.org/sys-v2/sys-core/service/go/pkg/coredb"

	"go.amplifyedge.org/sys-share-v2/sys-account/service/go/pkg"
	sharedAuth "go.amplifyedge.org/sys-share-v2/sys-account/service/go/pkg/shared"
)

func (ad *SysAccountRepo) allowNewAccount(ctx context.Context, in *pkg.AccountNewRequest) error {
	ad.log.Debugf("getting permission for new account creation")
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	// allow superadmin to create account
	if allowed := sharedAuth.IsSuperadmin(curAcc.Role); allowed {
		return nil
	}
	if len(in.Roles) > 0 {
		if in.Roles[0].OrgID != "" && in.Roles[0].ProjectID == "" {
			// allow org admin
			ad.log.Debugf("expecting org admin of: %s", in.Roles[0].OrgID)
			allowed, err := sharedAuth.AllowOrgAdmin(curAcc, in.Roles[0].OrgID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		} else if (in.Roles[0].OrgID == "" && in.Roles[0].ProjectID != "") || (in.Roles[0].OrgID != "" && in.Roles[0].ProjectID != "") {
			// allow project admin
			ad.log.Debugf("expecting project admin of org: %s, project: %s", in.Roles[0].OrgID)
			allowed, err := sharedAuth.AllowProjectAdmin(curAcc, "", in.Roles[0].ProjectID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		} else if in.Roles[0].OrgID == "" && in.Roles[0].ProjectID == "" {
			ad.log.Debugf("expecting superadmin")
			if allowed := sharedAuth.IsSuperadmin(in.Roles); !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		}
	}
	ad.log.Debugf("no match for current user, denying new account privilege")
	// disallow others
	return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
}

func hasOrgIds(in *pkg.Account) bool {
	for _, r := range in.Role {
		if r.OrgID != "" {
			return true
		}
	}
	return false
}

func hasProjectIds(in *pkg.Account) bool {
	for _, r := range in.Role {
		if r.ProjectID != "" {
			return true
		}
	}
	return false
}

func (ad *SysAccountRepo) allowGetAccount(ctx context.Context, idRequest *pkg.IdRequest) (*pkg.Account, error) {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	in, err := ad.getAccountAndRole(ctx, idRequest.Id, idRequest.Name)
	if err != nil {
		return nil, err
	}
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return in, nil
	}
	if allowed := sharedAuth.AllowSelf(curAcc, in.Id); allowed {
		return in, nil
	}
	if hasOrgIds(curAcc) && !hasProjectIds(curAcc) {
		for _, r := range in.Role {
			allowed, err := sharedAuth.AllowOrgAdmin(curAcc, r.OrgID)
			if allowed && err != nil {
				return in, nil
			}
		}
		return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	} else if hasOrgIds(curAcc) && hasProjectIds(curAcc) {
		for _, r := range in.Role {
			allowed, err := sharedAuth.AllowProjectAdmin(curAcc, r.OrgID, r.ProjectID)
			if allowed && err != nil {
				return in, nil
			}
		}
		return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}

	return in, nil
}

// authz for list accounts
func (ad *SysAccountRepo) allowListAccount(ctx context.Context) (*coresvc.QueryParams, error) {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	// allow all if it's superadmin
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return &coresvc.QueryParams{Params: map[string]interface{}{}}, nil
	}
	isAdm, idx := sharedAuth.IsAdmin(curAcc.Role)
	if isAdm {
		params := map[string]interface{}{}
		if curAcc.Role[*idx].OrgID != "" {
			params["org_id"] = curAcc.Role[*idx].OrgID
			// only allow org admin to query its own org
			return &coresvc.QueryParams{Params: params}, nil
		} else if curAcc.Role[*idx].ProjectID != "" {
			params["org_id"] = curAcc.Role[*idx].OrgID
			params["project_id"] = curAcc.Role[*idx].ProjectID
			// only allow project admin to query its own project
			return &coresvc.QueryParams{Params: params}, nil
		} else {
			return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters, Err: err}.Error())
		}
	}
	return &coresvc.QueryParams{Params: map[string]interface{}{"id": curAcc.Id}}, nil
}

func (ad *SysAccountRepo) allowAssignToRole(ctx context.Context, in *pkg.AssignAccountToRoleRequest) error {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	// allow superadmin to do anything
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return nil
	}
	isAdm, _ := sharedAuth.IsAdmin(curAcc.Role)
	if isAdm {
		if in.Role.OrgID != "" && in.Role.ProjectID == "" {
			// Org Admin
			allowed, err := sharedAuth.AllowOrgAdmin(curAcc, in.Role.OrgID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		} else if in.Role.ProjectID != "" {
			// Project Admin
			allowed, err := sharedAuth.AllowProjectAdmin(curAcc, in.Role.OrgID, in.Role.ProjectID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		}
		return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	// Allow self update account
	if sharedAuth.AllowSelf(curAcc, in.AssignedAccountId) {
		return nil
	}
	return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
}

// only allow superadmin to create new org.
func (ad *SysAccountRepo) allowNewOrg(ctx context.Context) error {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return nil
	}
	return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
}

// only allow superadmin and specified org admins to edit / update / delete org
func (ad *SysAccountRepo) allowUpdateDeleteOrg(ctx context.Context, orgId string) error {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return nil
	}
	// ORG ADMIN
	isAdm, _ := sharedAuth.IsAdmin(curAcc.Role)
	if !isAdm {
		return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	match, err := sharedAuth.AllowOrgAdmin(curAcc, orgId)
	if err != nil {
		return err
	}
	if !match {
		return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	return nil
}

// only allow org admin or superadmin to create new project.
func (ad *SysAccountRepo) allowNewProject(ctx context.Context, orgId string) error {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return nil
	}
	// ORG ADMIN
	isAdm, _ := sharedAuth.IsAdmin(curAcc.Role)
	if isAdm {
		match, err := sharedAuth.AllowOrgAdmin(curAcc, orgId)
		if err != nil {
			return err
		}
		if !match {
			return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
		}
		return nil
	}
	return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
}

// only allow project admin, org admin, or superadmin to edit / update / delete the project
func (ad *SysAccountRepo) allowUpdateDeleteProject(ctx context.Context, orgId string, projectId string) error {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return nil
	}
	isAdm, _ := sharedAuth.IsAdmin(curAcc.Role)
	if isAdm {
		// ORG ADMIN
		if orgId != "" && projectId == "" {
			match, err := sharedAuth.AllowOrgAdmin(curAcc, orgId)
			if err != nil {
				return err
			}
			if !match {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		}
		// PROJECT ADMIN
		if orgId != "" && projectId != "" {
			match, err := sharedAuth.AllowProjectAdmin(curAcc, orgId, projectId)
			if err != nil {
				return err
			}
			if !match {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		}
	}
	return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
}
