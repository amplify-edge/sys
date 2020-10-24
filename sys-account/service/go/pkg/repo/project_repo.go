package repo

import (
	"context"
	"fmt"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/getcouragenow/sys-share/sys-account/service/go/pkg/shared"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

func (ad *SysAccountRepo) projectFetchOrg(req *dao.Project) (*pkg.Project, error) {
	org, err := ad.store.GetOrg(&coresvc.QueryParams{Params: map[string]interface{}{"id": req.OrgId}})
	if err != nil {
		return nil, err
	}
	pkgOrg, err := org.ToPkgOrg(nil)
	if err != nil {
		return nil, err
	}
	return req.ToPkgProject(pkgOrg)
}

func (ad *SysAccountRepo) NewProject(ctx context.Context, in *pkg.ProjectRequest) (*pkg.Project, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot insert project: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	req, err := ad.store.FromPkgProject(in)
	if err != nil {
		return nil, err
	}
	if err := ad.store.InsertProject(req); err != nil {
		return nil, err
	}
	proj, err := ad.store.GetProject(&coresvc.QueryParams{Params: map[string]interface{}{"id": req.Id}})
	if err != nil {
		return nil, err
	}
	return ad.projectFetchOrg(proj)
}

func (ad *SysAccountRepo) GetProject(ctx context.Context, in *pkg.IdRequest) (*pkg.Project, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot get project: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	proj, err := ad.store.GetProject(&coresvc.QueryParams{Params: map[string]interface{}{"id": in.Id}})
	if err != nil {
		return nil, err
	}
	return ad.projectFetchOrg(proj)
}

func (ad *SysAccountRepo) ListProject(ctx context.Context, in *pkg.ListRequest) (*pkg.ListResponse, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot list project: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	var limit, cursor int64
	orderBy := in.OrderBy
	var err error
	filter := &coresvc.QueryParams{Params: map[string]interface{}{}}
	if in.IsDescending {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}
	cursor, err = ad.getCursor(in.CurrentPageId)
	if err != nil {
		return nil, err
	}
	if limit == 0 {
		limit = dao.DefaultLimit
	}
	projects, next, err := ad.store.ListProject(filter, orderBy, limit, cursor)
	var pkgProjects []*pkg.Project
	for _, p := range projects {
		pkgProject, err := ad.projectFetchOrg(p)
		if err != nil {
			return nil, err
		}
		pkgProjects = append(pkgProjects, pkgProject)
	}
	return &pkg.ListResponse{
		Projects:   pkgProjects,
		NextPageId: fmt.Sprintf("%d", next),
	}, nil
}

func (ad *SysAccountRepo) UpdateProject(ctx context.Context, in *pkg.ProjectRequest) (*pkg.Project, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot list project: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	req, err := ad.store.FromPkgProject(in)
	if err != nil {
		return nil, err
	}
	err = ad.store.UpdateProject(req)
	if err != nil {
		return nil, err
	}
	proj, err := ad.store.GetProject(&coresvc.QueryParams{Params: map[string]interface{}{"id": req.Id}})
	if err != nil {
		return nil, err
	}
	return ad.projectFetchOrg(proj)
}

func (ad *SysAccountRepo) DeleteProject(ctx context.Context, in *pkg.IdRequest) (*emptypb.Empty, error) {
	if in == nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot list project: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters})
	}
	err := ad.store.DeleteProject(in.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
