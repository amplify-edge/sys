package dao

import (
	"encoding/json"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	rpc "github.com/getcouragenow/sys-share/sys-account/server/rpc/v2"
	"github.com/getcouragenow/sys/sys-account/server/pkg/crud"
	log "github.com/sirupsen/logrus"
	"strconv"
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

func (p *Permission) ToProto() (*rpc.UserRoles, error) {
	role, err := strconv.Atoi(p.Role)
	if err != nil {
		return nil, err
	}
	if p.OrgId != "" {
		return &rpc.UserRoles{
			Role:     rpc.Roles(role),
			Resource: &rpc.UserRoles_Org{Org: &rpc.Org{Id: p.OrgId}},
		}, nil
	} else if p.ProjectId != "" {
		return &rpc.UserRoles{
			Role:     rpc.Roles(role),
			Resource: &rpc.UserRoles_Project{Project: &rpc.Project{Id: p.ProjectId}},
		}, nil
	} else if rpc.Roles(role) == rpc.Roles_SUPERADMIN {
		return &rpc.UserRoles{
			Role:     rpc.Roles_SUPERADMIN,
			Resource: nil,
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
