package repo

import (
	"context"
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
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	if in.Role.OrgID != "" && in.Role.ProjectID == "" {
		allowed, err := sharedAuth.AllowOrgAdmin(curAcc, in.Role.OrgID)
		if err != nil || !allowed {
			return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
		}
	} else if in.Role.OrgID == "" && in.Role.ProjectID != "" {
		allowed, err := sharedAuth.AllowProjectAdmin(curAcc, "", in.Role.ProjectID)
		if err != nil || !allowed {
			return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
		}
	} else if in.Role.OrgID == "" && in.Role.ProjectID == "" {
		if allowed := sharedAuth.IsSuperadmin(in.Role); !allowed {
			return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
		}
	}
	return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
}

func (ad *SysAccountRepo) allowGetAccount(ctx context.Context, id string) (*pkg.Account, error) {
	_, curAcc, err := ad.accountFromClaims(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	in, err := ad.getAccountAndRole(id)
	if err != nil {
		return nil, err
	}
	if sharedAuth.IsSuperadmin(curAcc.Role) {
		return in, nil
	}
	if in.Role.OrgID != "" && in.Role.ProjectID == "" {
		allowed, err := sharedAuth.AllowOrgAdmin(curAcc, in.Role.OrgID)
		if err != nil || !allowed {
			return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
		}
	} else if in.Role.OrgID == "" && in.Role.ProjectID != "" {
		allowed, err := sharedAuth.AllowProjectAdmin(curAcc, "", in.Role.ProjectID)
		if err != nil || !allowed {
			return nil, status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
		}
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
	if sharedAuth.IsAdmin(curAcc.Role) {
		params := map[string]interface{}{}
		if curAcc.Role.OrgID != "" {
			params["org_id"] = curAcc.Role.OrgID
			return &coresvc.QueryParams{Params: params}, nil
		} else if curAcc.Role.ProjectID != "" {
			params["org_id"] = curAcc.Role.OrgID
			params["project_id"] = curAcc.Role.ProjectID
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
	if sharedAuth.IsAdmin(curAcc.Role) {
		if in.Role.OrgID != "" && in.Role.ProjectID == "" {
			allowed, err := sharedAuth.AllowOrgAdmin(curAcc, in.Role.OrgID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
		} else if in.Role.ProjectID != "" {
			allowed, err := sharedAuth.AllowProjectAdmin(curAcc, in.Role.OrgID, in.Role.ProjectID)
			if err != nil || !allowed {
				return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
			}
		}
		return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
	}
	return status.Errorf(codes.PermissionDenied, sharedAuth.Error{Reason: sharedAuth.ErrRequestUnauthenticated, Err: err}.Error())
}
