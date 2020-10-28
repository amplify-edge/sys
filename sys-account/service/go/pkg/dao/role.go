package dao

import (
	"errors"
	"fmt"
	utilities "github.com/getcouragenow/sys-share/sys-core/service/config"
	log "github.com/sirupsen/logrus"
	"time"

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

func (a *AccountDB) FromPkgRoleRequest(role *pkg.UserRoles, accountId string) *Role {
	return &Role{
		ID:        utilities.NewID(),
		AccountId: accountId,
		Role:      int(role.Role),
		ProjectId: role.ProjectID,
		OrgId:     role.OrgID,
		CreatedAt: time.Now().UTC().Unix(),
	}
}

func (a *AccountDB) FetchRoles(accountId string) ([]*Role, error) {
	queryParam := &coresvc.QueryParams{Params: map[string]interface{}{
		"account_id": accountId,
	}}
	a.log.Debugf("Role query param: %v", queryParam.Params)
	listRoles, err := a.ListRole(queryParam)
	if err != nil {
		return nil, err
	}
	return listRoles, nil
}

func (p *Role) ToPkgRole() (*pkg.UserRoles, error) {
	role := p.Role
	if role == int(pkg.INVALID) || role > int(pkg.SUPERADMIN) {
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

func roleToQueryParam(acc *Role) (res coresvc.QueryParams, err error) {
	return coresvc.AnyToQueryParam(acc, true)
}

// CreateSQL will only be called once by sys-core see sys-core API.
func (p Role) CreateSQL() []string {
	fields := initFields(RolesColumns, RolesColumnsType)
	tbl := coresvc.NewTable(RolesTableName, fields, []string{})
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
	if err != nil {
		a.log.Debugf("Unable to scan role to struct: %v", err)
		return nil, err
	}
	return &p, nil
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
	if p.OrgId != "" {
		_, err := a.GetOrg(&coresvc.QueryParams{Params: map[string]interface{}{"id": p.OrgId}})
		if err != nil {
			return err
		}
	}
	if p.ProjectId != "" {
		_, err := a.GetProject(&coresvc.QueryParams{Params: map[string]interface{}{"id": p.ProjectId}})
		if err != nil {
			return err
		}
	}
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
	delete(filterParam.Params, "id")
	delete(filterParam.Params, "account_id")
	stmt, args, err := sq.Update(RolesTableName).SetMap(filterParam.Params).
		Where(sq.Eq{"id": p.ID}).ToSql()
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
