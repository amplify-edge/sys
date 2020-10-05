package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getcouragenow/sys-share/pkg"
	log "github.com/sirupsen/logrus"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/crud"
)

type Permission struct {
	ID        string
	AccountId string
	Role      string
	ProjectId string
	OrgId     string
	CreatedAt int64 // UTC Unix timestamp
	UpdatedAt int64
}

func (a *AccountDB) FromPkgRole(role *pkg.UserRoles, accountId string) (*Permission, error) {
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

func (p *Permission) ToPkgRole() (*pkg.UserRoles, error) {
	role, err := strconv.Atoi(p.Role)
	if err != nil {
		return nil, err
	}
	if p.OrgId != "" {
		return &pkg.UserRoles{
			Role:  pkg.Roles(role),
			OrgID: p.OrgId,
		}, nil
	} else if p.ProjectId != "" {
		return &pkg.UserRoles{
			Role:      pkg.Roles(role),
			ProjectID: p.ProjectId,
		}, nil
	} else if pkg.Roles(role) == 4 {
		return &pkg.UserRoles{
			Role: pkg.Roles(role),
			All:  true,
		}, nil
	} else if pkg.Roles(role) == 1 {
		return &pkg.UserRoles{
			Role: pkg.Roles(role),
			All:  false,
		}, nil
	}
	return nil, errors.New("invalid Role")
}

func (p Permission) TableName() string {
	return tableName(PermTableName, "_")
}

func permissionToQueryParam(acc *Permission) (res QueryParams, err error) {
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
func (p Permission) CreateSQL() []string {
	fields := initFields(PermColumns)
	tbl := crud.NewTable(PermTableName, fields)
	return []string{tbl.CreateTable()}
}

func (a *AccountDB) getPermSelectStatements(aqp *QueryParams) (string, []interface{}, error) {
	baseStmt := sq.Select(PermColumns).From(PermTableName)
	if aqp != nil && aqp.Params != nil {
		for k, v := range aqp.Params {
			baseStmt = baseStmt.Where(sq.Eq{k: v})
		}
	}
	return baseStmt.ToSql()
}

func (a *AccountDB) GetRole(aqp *QueryParams) (*Permission, error) {
	var p Permission
	selectStmt, args, err := a.getPermSelectStatements(aqp)
	if err != nil {
		return nil, err
	}
	doc, err := a.QueryOne(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = document.StructScan(doc, &p)
	return &p, err
}

func (a *AccountDB) ListRole(aqp *QueryParams) ([]*Permission, error) {
	var perms []*Permission
	selectStmt, args, err := a.getPermSelectStatements(aqp)
	if err != nil {
		return nil, err
	}
	res, err := a.Query(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = res.Iterate(func(d document.Document) error {
		var perm Permission
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

func (a *AccountDB) InsertRole(p *Permission) error {
	var allVals [][]interface{}
	aqp, err := permissionToQueryParam(p)
	if err != nil {
		return err
	}
	log.Printf("query params: %v", aqp)
	columns, values := aqp.ColumnsAndValues()
	stmt, args, err := sq.Insert(PermTableName).
		Columns(columns...).
		Values(values...).
		ToSql()
	a.log.WithFields(log.Fields{
		"statement": stmt,
		"args":      args,
	})
	if err != nil {
		return err
	}
	allVals = append(allVals, args)
	return a.Exec([]string{stmt}, allVals)
}

func (a *AccountDB) UpdateRole(p *Permission) error {
	aqp, err := permissionToQueryParam(p)
	if err != nil {
		return err
	}
	var values [][]interface{}
	stmt, args, err := sq.Update(PermTableName).SetMap(aqp.Params).ToSql()
	if err != nil {
		return err
	}
	values = append(values, args)
	return a.Exec([]string{stmt}, values)
}

func (a *AccountDB) DeleteRole(id string) error {
	var values [][]interface{}
	stmt, args, err := sq.Delete(PermTableName).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	values = append(values, args)
	return a.Exec([]string{stmt}, values)
}
