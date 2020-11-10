package dao

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

type Org struct {
	Id             string `genji:"id" json:"id,omitempty" coredb:"primary"`
	Name           string `genji:"name" json:"name,omitempty"`
	LogoResourceId string `genji:"logo_resource_id" json:"logo_resource_id,omitempty"`
	Contact        string `genji:"contact" json:"contact,omitempty"`
	CreatedAt      int64  `genji:"created_at" json:"created_at"`
	AccountId      string `genji:"account_id" json:"account_id"`
}

var (
	orgUniqueIndex     = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_name ON %s(name)", OrgTableName, OrgTableName)
	orgLogoUniqueIndex = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_logo_resource_id ON %s(logo_resource_id)", OrgTableName, OrgTableName)
)

func (a *AccountDB) FromPkgOrgRequest(org *pkg.OrgRequest, id string) (*Org, error) {
	orgId := id
	if orgId == "" {
		orgId = coresvc.NewID()
	}
	return &Org{
		Id:             orgId,
		Name:           org.Name,
		LogoResourceId: org.LogoFilepath,
		Contact:        org.Contact,
		CreatedAt:      coresvc.CurrentTimestamp(),
		AccountId:      org.CreatorId,
	}, nil
}

func (a *AccountDB) FromPkgOrgRequestToQueryFilter(org *pkg.OrgRequest) (*coresvc.QueryParams, error) {
	qf, err := coresvc.AnyToQueryParam(org, false)
	if err != nil {
		return nil, err
	}
	delete(qf.Params, "logo_upload_bytes")
	return &qf, nil
}

func (o *Org) ToPkgOrg(projects []*pkg.Project, logo []byte) (*pkg.Org, error) {
	return &pkg.Org{
		Id:             o.Id,
		Name:           o.Name,
		LogoResourceId: o.LogoResourceId,
		Logo:           logo,
		Contact:        o.Contact,
		CreatedAt:      o.CreatedAt,
		CreatorId:      o.AccountId,
		Projects:       projects,
	}, nil
}

func (o Org) CreateSQL() []string {
	fields := coresvc.GetStructTags(o)
	// tbl := coresvc.NewTable(OrgTableName, fields, []string{orgUniqueIndex})
	tbl := coresvc.NewTable(OrgTableName, fields, []string{orgUniqueIndex, orgLogoUniqueIndex})
	return tbl.CreateTable()
}

func orgToQueryParam(org *Org) (res coresvc.QueryParams, err error) {
	return coresvc.AnyToQueryParam(org, true)
}

func (a *AccountDB) orgQueryFilter(filter *coresvc.QueryParams) sq.SelectBuilder {
	baseStmt := sq.Select(a.orgColumns).From(OrgTableName)
	if filter != nil && filter.Params != nil {
		for k, v := range filter.Params {
			baseStmt = baseStmt.Where(sq.Eq{k: v})
		}
	}
	return baseStmt
}

func (a *AccountDB) GetOrg(filterParam *coresvc.QueryParams) (*Org, error) {
	var o Org
	selectStmt, args, err := a.orgQueryFilter(filterParam).ToSql()
	if err != nil {
		return nil, err
	}
	a.log.WithFields(log.Fields{
		"queryStatement": selectStmt,
		"arguments":      args,
	}).Debug("Querying org")
	doc, err := a.db.QueryOne(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = doc.StructScan(&o)
	return &o, err
}

func (a *AccountDB) ListOrg(filterParam *coresvc.QueryParams, orderBy string, limit, cursor int64) ([]*Org, int64, error) {
	var orgs []*Org
	baseStmt := a.orgQueryFilter(filterParam)
	selectStmt, args, err := a.listSelectStatements(baseStmt, orderBy, limit, &cursor)
	if err != nil {
		return nil, 0, err
	}
	res, err := a.db.Query(selectStmt, args...)
	if err != nil {
		return nil, 0, err
	}
	err = res.Iterate(func(d document.Document) error {
		var org Org
		if err = document.StructScan(d, &org); err != nil {
			return err
		}
		orgs = append(orgs, &org)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	res.Close()
	return orgs, orgs[len(orgs)-1].CreatedAt, nil
}

func (a *AccountDB) InsertOrg(o *Org) error {
	filterParam, err := orgToQueryParam(o)
	if err != nil {
		return err
	}
	columns, values := filterParam.ColumnsAndValues()
	if len(columns) != len(values) {
		return fmt.Errorf("error: length mismatch: cols: %d, vals: %d", len(columns), len(values))
	}
	stmt, args, err := sq.Insert(OrgTableName).
		Columns(columns...).
		Values(values...).
		ToSql()
	a.log.WithFields(log.Fields{
		"statement": stmt,
		"args":      args,
	}).Debugf("insert to %s table", OrgTableName)
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) UpdateOrg(o *Org) error {
	filterParam, err := orgToQueryParam(o)
	if err != nil {
		return err
	}
	a.log.Debugf("org update param: %v", filterParam)
	delete(filterParam.Params, "id")
	stmt, args, err := sq.Update(OrgTableName).SetMap(filterParam.Params).
		Where(sq.Eq{"id": o.Id}).ToSql()
	if err != nil {
		return err
	}
	a.log.Debugf(
		"update org statement: %v, args: %v", stmt,
		args,
	)
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) DeleteOrg(id string) error {
	stmt, args, err := sq.Delete(OrgTableName).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	pstmt, pargs, err := sq.Delete(ProjectTableName).Where("org_id = ?", id).ToSql()
	if err != nil {
		return err
	}
	return a.db.BulkExec(map[string][]interface{}{
		stmt:  args,
		pstmt: pargs,
	})
}
