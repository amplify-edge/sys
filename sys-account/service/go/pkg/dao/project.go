package dao

import (
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

type Project struct {
	Id        string `genji:"id" coredb:"primary"`
	Name      string `genji:"name"`
	LogoUrl   string `genji:"logo_url"`
	CreatedAt int64  `genji:"created_at"`
	AccountId string `genji:"account_id"`
	OrgId     string `genji:"org_id"`
}

var (
	projectUniqueIndex = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_name ON %s(name)", ProjectTableName, ProjectTableName)
)

func (a *AccountDB) FromPkgProject(p *pkg.ProjectRequest) (*Project, error) {
	var orgId string
	if p.OrgId == "" {
		return nil, errors.New("project organization id required")
	}
	if p.OrgId != "" {
		orgId = p.OrgId
	}
	return &Project{
		Id:        coresvc.NewID(),
		Name:      p.Name,
		LogoUrl:   p.LogoUrl,
		CreatedAt: coresvc.CurrentTimestamp(),
		AccountId: p.CreatorId,
		OrgId:     orgId,
	}, nil
}

func (p *Project) ToPkgProject(org *pkg.Org) (*pkg.Project, error) {
	return &pkg.Project{
		Id:        p.Id,
		Name:      p.Name,
		LogoUrl:   p.LogoUrl,
		CreatedAt: p.CreatedAt,
		CreatorId: p.AccountId,
		OrgId:     p.OrgId,
		Org:       org,
	}, nil
}

func (p Project) CreateSQL() []string {
	fields := coresvc.GetStructTags(p)
	tbl := coresvc.NewTable(ProjectTableName, fields, []string{projectUniqueIndex})
	return tbl.CreateTable()
}

func projectToQueryParam(p *Project) (res coresvc.QueryParams, err error) {
	return coresvc.AnyToQueryParam(p, true)
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

func (a *AccountDB) GetProject(filterParam *coresvc.QueryParams) (*Project, error) {
	var p Project
	selectStmt, args, err := a.projectQueryFilter(filterParam).ToSql()
	if err != nil {
		return nil, err
	}
	a.log.WithFields(log.Fields{
		"queryStatement": selectStmt,
		"arguments":      args,
	}).Debug("Querying roles")
	doc, err := a.db.QueryOne(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = doc.StructScan(&p)
	return &p, err
}

func (a *AccountDB) ListProject(filterParam *coresvc.QueryParams, orderBy string, limit, cursor int64) ([]*Project, int64, error) {
	var projs []*Project
	baseStmt := a.projectQueryFilter(filterParam)
	selectStmt, args, err := a.listSelectStatements(baseStmt, orderBy, limit, &cursor)
	if err != nil {
		return nil, 0, err
	}
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
