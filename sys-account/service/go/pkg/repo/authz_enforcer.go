package repo

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"
)

func (ad *SysAccountRepo) allowNewAccount(ctx context.Context, in *pkg.AccountNewRequest) error {
	ad.log.Debugf("getting permission for new account creation")
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if allowed := sharedAuth.IsSuperadmin(curAcc.Role); allowed {
		return nil
	}
	if len(in.Roles) > 0 {
		if in.Roles[0].OrgID != "" && in.Roles[0].ProjectID == "" {
			ad.log.Debugf("expecting org admin of: %s", in.Roles[0].OrgID)
			allowed, err := sharedAuth.AllowOrgAdmin(curAcc, in.Roles[0].OrgID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
			return nil
		} else if (in.Roles[0].OrgID == "" && in.Roles[0].ProjectID != "") || (in.Roles[0].OrgID != "" && in.Roles[0].ProjectID != "") {
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
	in, err := ad.getAccountAndRole(idRequest.Id, idRequest.Name)
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
	Email          string `json:"string"`
	Password       string `json:"password"`
	AvatarFilePath string `json:"avatar_filepath"`
	AvatarBytes    []byte `json:"avatar_bytes"`
}

// Initial User Creation via CLI only
func (ad *SysAccountRepo) InitSuperUser(in *SuperAccountRequest) error {
	if in == nil {
		return fmt.Errorf("error unable to proceed, user is nil")
	}
	avatar, err := ad.frepo.UploadFile(in.AvatarFilePath, in.AvatarBytes)
	if err != nil {
		return err
	}
	newAcc := &pkg.AccountNewRequest{
		Email:          in.Email,
		Password:       in.Password,
		Roles:          []*pkg.UserRoles{{Role: pkg.SUPERADMIN}},
		AvatarFilepath: avatar.GetResourceId(),
	}
	_, err = ad.store.InsertFromPkgAccountRequest(newAcc, true)
	if err != nil {
		ad.log.Debugf("error unable to create super-account request: %v", err)
		return err
	}
	return nil
}
