package dao

import (
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

type Project struct {
	Id             string `json:"id" genji:"id" coredb:"primary"`
	Name           string `json:"name,omitempty" genji:"name"`
	LogoResourceId string `json:"logo_resource_id" genji:"logo_resource_id"`
	CreatedAt      int64  `json:"created_at" genji:"created_at"`
	AccountId      string `json:"account_id" genji:"account_id"`
	OrgId          string `json:"org_id" genji:"org_id"`
	OrgName        string `json:"org_name" genji:"org_name"`
}

var (
	projectUniqueIndex     = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_name ON %s(name)", ProjectTableName, ProjectTableName)
	projectLogoUniqueIndex = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_logo_resource_id ON %s(logo_resource_id)", ProjectTableName, ProjectTableName)
)

func (a *AccountDB) FromPkgProject(p *pkg.ProjectRequest) (*Project, error) {
	var orgId, orgName string
	if p.OrgId == "" && p.OrgName == "" {
		return nil, errors.New("project organization id required")
	}
	if p.OrgId != "" {
		orgId = p.OrgId
	}
	if p.OrgName != "" {
		orgName = p.OrgName
	}
	return &Project{
		Id:             sharedConfig.NewID(),
		Name:           p.Name,
		LogoResourceId: p.LogoFilepath,
		CreatedAt:      sharedConfig.CurrentTimestamp(),
		AccountId:      p.CreatorId,
		OrgId:          orgId,
		OrgName:        orgName,
	}, nil
}

func (p *Project) ToPkgProject(org *pkg.Org, logo []byte) (*pkg.Project, error) {
	porg := &pkg.Org{}
	if org != nil {
		porg = org
	}
	return &pkg.Project{
		Id:             p.Id,
		Name:           p.Name,
		LogoResourceId: p.LogoResourceId,
		Logo:           logo,
		CreatedAt:      p.CreatedAt,
		CreatorId:      p.AccountId,
		OrgId:          p.OrgId,
		OrgName:        p.OrgName,
		Org:            porg,
	}, nil
}

func (p Project) CreateSQL() []string {
	fields := coresvc.GetStructTags(p)
	tbl := coresvc.NewTable(ProjectTableName, fields, []string{projectUniqueIndex, projectLogoUniqueIndex})
	return tbl.CreateTable()
}

func projectToQueryParam(p *Project) (res coresvc.QueryParams, err error) {
	qf, err := coresvc.AnyToQueryParam(p, true)
	if err != nil {
		return coresvc.QueryParams{}, err
	}
	return qf, nil
}

func (a *AccountDB) projectQueryFilter(filter *coresvc.QueryParams) sq.SelectBuilder {
	baseStmt := sq.Select(a.projectColumns).From(ProjectTableName)
	if filter != nil && filter.Params != nil {
		for k, v := range filter.Params {
			baseStmt = baseStmt.Where(sq.Eq{k: v})
		}
	}
	return baseStmt
}

func (a *AccountDB) projectLikeFilter(filter *coresvc.QueryParams) sq.SelectBuilder {
	baseStmt := sq.Select(a.projectColumns).From(ProjectTableName)
	if filter != nil && filter.Params != nil {
		for k, v := range filter.Params {
			baseStmt = baseStmt.Where(sq.Like{k: "%" + v.(string) + "%"})
		}
	}
	return baseStmt
}

func (a *AccountDB) GetProject(filterParam *coresvc.QueryParams) (*Project, error) {
	var p Project
	selectStmt, args, err := a.projectQueryFilter(filterParam).ToSql()
	if err != nil {
		return nil, err
	}
	a.log.WithFields(log.Fields{
		"queryStatement": selectStmt,
		"arguments":      args,
	}).Debug("Querying projects")
	doc, err := a.db.QueryOne(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = doc.StructScan(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (a *AccountDB) ListProject(filterParam *coresvc.QueryParams, orderBy string, limit, cursor int64) ([]*Project, int64, error) {
	var projs []*Project
	baseStmt := a.projectLikeFilter(filterParam)
	selectStmt, args, err := a.listSelectStatements(baseStmt, orderBy, limit, &cursor)
	if err != nil {
		return nil, 0, err
	}
	a.log.WithFields(log.Fields{
		"queryStatement": selectStmt,
		"arguments":      args,
	}).Debug("List projects")
	res, err := a.db.Query(selectStmt, args...)
	if err != nil {
		return nil, 0, err
	}
	err = res.Iterate(func(d document.Document) error {
		var p Project
		if err = document.StructScan(d, &p); err != nil {
			return err
		}
		projs = append(projs, &p)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	if len(projs) > 0 {
		return projs, projs[len(projs)-1].CreatedAt, nil
	}
	res.Close()
	return projs, 0, nil
}

func (a *AccountDB) InsertProject(p *Project) error {
	_, err := a.GetOrg(&coresvc.QueryParams{Params: map[string]interface{}{"id": p.OrgId}})
	if err != nil {
		return err
	}
	filterParam, err := projectToQueryParam(p)
	if err != nil {
		return err
	}
	columns, values := filterParam.ColumnsAndValues()
	if len(columns) != len(values) {
		return fmt.Errorf("error: length mismatch: cols: %d, vals: %d", len(columns), len(values))
	}
	a.log.WithFields(log.Fields{
		"columns": columns,
		"values":  values,
	}).Debug("insert into projects table")
	stmt, args, err := sq.Insert(ProjectTableName).
		Columns(columns...).
		Values(values...).
		ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) UpdateProject(p *Project) error {
	filterParam, err := projectToQueryParam(p)
	if err != nil {
		return err
	}
	delete(filterParam.Params, "id")
	delete(filterParam.Params, "org_id")
	delete(filterParam.Params, "org_name")
	stmt, args, err := sq.Update(ProjectTableName).SetMap(filterParam.Params).
		Where(sq.Eq{"id": p.Id}).ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) DeleteProject(id string) error {
	stmt, args, err := sq.Delete(ProjectTableName).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}
