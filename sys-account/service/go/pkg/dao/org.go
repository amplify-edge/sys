package dao

import (
	"encoding/json"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

type Org struct {
	Id        string `genji:"id"`
	Name      string `genji:"name"`
	LogoUrl   string `genji:"logo_url"`
	Contact   string `genji:"contact"`
	CreatedAt int64  `genji:"created_at"`
	AccountId string `genji:"account_id"`
}

func (a *AccountDB) FromPkgOrg(org *pkg.OrgRequest) (*Org, error) {
	return &Org{
		Id:        coresvc.NewID(),
		Name:      org.Name,
		LogoUrl:   org.LogoUrl,
		Contact:   org.Contact,
		CreatedAt: coresvc.CurrentTimestamp(),
		AccountId: org.CreatorId,
	}, nil
}

func (o *Org) ToPkgOrg(projects []*pkg.Project) (*pkg.Org, error) {
	return &pkg.Org{
		Id:        o.Id,
		Name:      o.Name,
		LogoUrl:   o.LogoUrl,
		Contact:   o.Contact,
		CreatedAt: o.CreatedAt,
		CreatorId: o.AccountId,
		Projects:  projects,
	}, nil
}

func (o Org) CreateSQL() []string {
	fields := initFields(OrgColumns, OrgColumnsType)
	// tbl := coresvc.NewTable(OrgTableName, fields, []string{orgUniqueIndex})
	tbl := coresvc.NewTable(OrgTableName, fields, []string{})
	return tbl.CreateTable()
}

func orgToQueryParam(org *Org) (res coresvc.QueryParams, err error) {
	jstring, err := json.Marshal(org)
	if err != nil {
		return coresvc.QueryParams{}, err
	}
	var params map[string]interface{}
	err = json.Unmarshal(jstring, &params)
	res.Params = params
	return res, err
}

func (a *AccountDB) orgQueryFilter(filter *coresvc.QueryParams) sq.SelectBuilder {
	baseStmt := sq.Select(OrgColumns).From(OrgTableName)
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
	}).Debug("Querying roles")
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
	stmt, args, err := sq.Update(RolesTableName).SetMap(filterParam.Params).ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) DeleteOrg(id string) error {
	stmt, args, err := sq.Delete(OrgTableName).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}
