package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"

	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
)

type Role struct {
	ID        string `genji:"id"`
	AccountId string `genji:"account_id"`
	Role      int    `genji:"role"`
	ProjectId string `genji:"project_id"`
	OrgId     string `genji:"org_id"`
	CreatedAt int64  `genji:"created_at"`
	UpdatedAt int64  `genji:"updated_at"`
}

var (
	rolesUniqueIndex = fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_account_id ON %s(account_id)`, RolesTableName)
)

func (a *AccountDB) FromPkgRole(role *pkg.UserRoles, accountId string) (*Role, error) {
	queryParam := &coresvc.QueryParams{Params: map[string]interface{}{
		"account_id": accountId,
	}}
	if role.Role > 0 && role.Role <= 4 { // Guest, Member, Admin, or SuperAdmin
		queryParam.Params["role"] = fmt.Sprintf("%d", role.Role)
	}
	if role.OrgID != "" {
		queryParam.Params["org_id"] = role.OrgID
	}
	if role.ProjectID != "" {
		queryParam.Params["project_id"] = role.ProjectID
	}
	return a.GetRole(queryParam)
}

func (p *Role) ToPkgRole() (*pkg.UserRoles, error) {
	role := p.Role
	if role == 0 || role >= 4 {
		return nil, errors.New("invalid role")
	}
	userRole := &pkg.UserRoles{
		Role: pkg.Roles(role),
	}
	if p.OrgId != "" {
		userRole.OrgID = p.OrgId
	}
	if p.ProjectId != "" {
		userRole.ProjectID = p.ProjectId
	}
	if pkg.Roles(role) == 4 {
		userRole.All = true
	}
	if pkg.Roles(role) == 1 {
		userRole.All = false
	}
	return userRole, nil
}

func (p Role) TableName() string {
	return RolesTableName
}

func roleToQueryParam(acc *Role) (res coresvc.QueryParams, err error) {
	jstring, err := json.Marshal(acc)
	if err != nil {
		return coresvc.QueryParams{}, err
	}
	var params map[string]interface{}
	err = json.Unmarshal(jstring, &params)
	res.Params = params
	return res, err
}

// CreateSQL will only be called once by sys-core see sys-core API.
func (p Role) CreateSQL() []string {
	fields := initFields(RolesColumns, RolesColumnsType)
	tbl := coresvc.NewTable(RolesTableName, fields, []string{rolesUniqueIndex})
	return tbl.CreateTable()
}

func (a *AccountDB) getRolesSelectStatements(filterParam *coresvc.QueryParams) (string, []interface{}, error) {
	baseStmt := sq.Select(RolesColumns).From(RolesTableName)
	if filterParam != nil && filterParam.Params != nil {
		for k, v := range filterParam.Params {
			baseStmt = baseStmt.Where(sq.Eq{k: v})
		}
	}
	return baseStmt.ToSql()
}

func (a *AccountDB) GetRole(filterParam *coresvc.QueryParams) (*Role, error) {
	var p Role
	selectStmt, args, err := a.getRolesSelectStatements(filterParam)
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

func (a *AccountDB) ListRole(filterParam *coresvc.QueryParams) ([]*Role, error) {
	var perms []*Role
	selectStmt, args, err := a.getRolesSelectStatements(filterParam)
	if err != nil {
		return nil, err
	}
	res, err := a.db.Query(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = res.Iterate(func(d document.Document) error {
		var perm Role
		if err = document.StructScan(d, &perm); err != nil {
			return err
		}
		perms = append(perms, &perm)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (a *AccountDB) InsertRole(p *Role) error {
	filterParam, err := roleToQueryParam(p)
	if err != nil {
		return err
	}
	columns, values := filterParam.ColumnsAndValues()
	if len(columns) != len(values) {
		return fmt.Errorf("error: length mismatch: cols: %d, vals: %d", len(columns), len(values))
	}
	stmt, args, err := sq.Insert(RolesTableName).
		Columns(columns...).
		Values(values...).
		ToSql()
	a.log.WithFields(log.Fields{
		"statement": stmt,
		"args":      args,
	}).Debug("insert to roles table")
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) UpdateRole(p *Role) error {
	filterParam, err := roleToQueryParam(p)
	if err != nil {
		return err
	}
	stmt, args, err := sq.Update(RolesTableName).SetMap(filterParam.Params).ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) DeleteRole(id string) error {
	stmt, args, err := sq.Delete(RolesTableName).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}
