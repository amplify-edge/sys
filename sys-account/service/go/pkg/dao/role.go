package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"

	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/crud"
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

func (a *AccountDB) FromPkgRole(role *pkg.UserRoles, accountId string) (*Role, error) {
	queryParam := &QueryParams{Params: map[string]interface{}{
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
	return tableName(RolesTableName, "_")
}

func permissionToQueryParam(acc *Role) (res QueryParams, err error) {
	jstring, err := json.Marshal(acc)
	if err != nil {
		return QueryParams{}, err
	}
	var params map[string]interface{}
	err = json.Unmarshal(jstring, &params)
	res.Params = params
	return res, err
}

// CreateSQL will only be called once by sys-core see sys-core API.
func (p Role) CreateSQL() []string {
	fields := initFields(RolesColumns, RolesColumnsType)
	tbl := crud.NewTable(p.TableName(), fields)
	return []string{tbl.CreateTable(),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_account_id ON %s(account_id)`, p.TableName())}
}

func (a *AccountDB) getRolesSelectStatements(aqp *QueryParams) (string, []interface{}, error) {
	baseStmt := sq.Select(RolesColumns).From(tableName(RolesTableName, "_"))
	if aqp != nil && aqp.Params != nil {
		for k, v := range aqp.Params {
			baseStmt = baseStmt.Where(sq.Eq{k: v})
		}
	}
	return baseStmt.ToSql()
}

func (a *AccountDB) GetRole(aqp *QueryParams) (*Role, error) {
	var p Role
	selectStmt, args, err := a.getRolesSelectStatements(aqp)
	if err != nil {
		return nil, err
	}
	a.log.WithFields(log.Fields{
		"queryStatement": selectStmt,
		"arguments":      args,
	}).Debug("Querying roles")
	doc, err := a.QueryOne(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = document.StructScan(doc, &p)
	return &p, err
}

func (a *AccountDB) ListRole(aqp *QueryParams) ([]*Role, error) {
	var perms []*Role
	selectStmt, args, err := a.getRolesSelectStatements(aqp)
	if err != nil {
		return nil, err
	}
	res, err := a.Query(selectStmt, args...)
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
	aqp, err := permissionToQueryParam(p)
	if err != nil {
		return err
	}
	columns, values := aqp.ColumnsAndValues()
	stmt, args, err := sq.Insert(tableName(RolesTableName, "_")).
		Columns(columns...).
		Values(values...).
		ToSql()
	a.log.WithFields(log.Fields{
		"statement": stmt,
		"args":      args,
	}).Info("INSERT to permissions table")
	if err != nil {
		return err
	}
	return a.Exec(stmt, args)
}

func (a *AccountDB) UpdateRole(p *Role) error {
	aqp, err := permissionToQueryParam(p)
	if err != nil {
		return err
	}
	stmt, args, err := sq.Update(tableName(RolesTableName, "_")).SetMap(aqp.Params).ToSql()
	if err != nil {
		return err
	}
	return a.Exec(stmt, args)
}

func (a *AccountDB) DeleteRole(id string) error {
	var values [][]interface{}
	stmt, args, err := sq.Delete(tableName(RolesTableName, "_")).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	values = append(values, args)
	return a.Exec(stmt, args)
}
