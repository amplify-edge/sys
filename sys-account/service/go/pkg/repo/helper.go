package repo

import (
	"context"
	"fmt"
	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"
)

func timestampNow() int64 {
	return time.Now().UTC().Unix()
}

func (ad *SysAccountRepo) allowNewAccount(ctx context.Context, in *pkg.Account) error {
	ad.log.Debugf("getting permission for new account creation")
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if allowed := sharedAuth.IsSuperadmin(curAcc.Role); allowed {
		return nil
	}
	if len(in.Role) > 0 {
		if in.Role[0].OrgID != "" && in.Role[0].ProjectID == "" {
			ad.log.Debugf("expecting org admin of: %s", in.Role[0].OrgID)
			allowed, err := sharedAuth.AllowOrgAdmin(curAcc, in.Role[0].OrgID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		} else if (in.Role[0].OrgID == "" && in.Role[0].ProjectID != "") || (in.Role[0].OrgID != "" && in.Role[0].ProjectID != "") {
			ad.log.Debugf("expecting project admin of org: %s, project: %s", in.Role[0].OrgID)
			allowed, err := sharedAuth.AllowProjectAdmin(curAcc, "", in.Role[0].ProjectID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		} else if in.Role[0].OrgID == "" && in.Role[0].ProjectID == "" {
			ad.log.Debugf("expecting superadmin")
			if allowed := sharedAuth.IsSuperadmin(in.Role); !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		}
	}
	ad.log.Debugf("no match for current user, denying new account privilege")
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

func (ad *SysAccountRepo) allowGetAccount(ctx context.Context, id string) (*pkg.Account, error) {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	in, err := ad.getAccountAndRole(id, "")
	if err != nil {
		return nil, err
	}
	if sharedAuth.IsSuperadmin(curAcc.Role) {
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
	} else if !hasOrgIds(curAcc) && hasProjectIds(curAcc) {
		for _, r := range in.Role {
			allowed, err := sharedAuth.AllowProjectAdmin(curAcc, r.OrgID, r.ProjectID)
			if allowed && err != nil {
				return in, nil
			}
		}
		return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if allowed := sharedAuth.AllowSelf(curAcc, in.Id); !allowed {
		return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	return in, nil
}

func (ad *SysAccountRepo) allowListAccount(ctx context.Context) (*coresvc.QueryParams, error) {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return &coresvc.QueryParams{Params: map[string]interface{}{}}, nil
	}
	isAdm, idx := sharedAuth.IsAdmin(curAcc.Role)
	if isAdm {
		params := map[string]interface{}{}
		if curAcc.Role[*idx].OrgID != "" {
			params["org_id"] = curAcc.Role[*idx].OrgID
			return &coresvc.QueryParams{Params: params}, nil
		} else if curAcc.Role[*idx].ProjectID != "" {
			params["org_id"] = curAcc.Role[*idx].OrgID
			params["project_id"] = curAcc.Role[*idx].ProjectID
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
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return nil
	}
	isAdm, _ := sharedAuth.IsAdmin(curAcc.Role)
	if isAdm {
		if in.Role.OrgID != "" && in.Role.ProjectID == "" {
			allowed, err := sharedAuth.AllowOrgAdmin(curAcc, in.Role.OrgID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		} else if in.Role.ProjectID != "" {
			allowed, err := sharedAuth.AllowProjectAdmin(curAcc, in.Role.OrgID, in.Role.ProjectID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		}
		return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
}

type SuperAccountRequest struct {
	Email    string `json:"string"`
	Password string `json:"password"`
}

// Initial User Creation via CLI only
func (ad *SysAccountRepo) InitSuperUser(in *SuperAccountRequest) error {
	if in == nil {
		return fmt.Errorf("error unable to proceed, user is nil")
	}
	newAcc := &pkg.Account{
		Id:        sharedConfig.NewID(),
		Email:     in.Email,
		Password:  in.Password,
		Role:      []*pkg.UserRoles{{Role: pkg.SUPERADMIN, All: true}},
		CreatedAt: timestampNow(),
		UpdatedAt: timestampNow(),
		Disabled:  false,
		Verified:  true,
	}
	_, err := ad.store.InsertFromPkgAccountRequest(newAcc)
	if err != nil {
		ad.log.Debugf("error unable to create super-account request: %v", err)
		return err
	}
	return nil
}
