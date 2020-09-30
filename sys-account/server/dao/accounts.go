package dao

import (
	"encoding/json"
	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	"github.com/getcouragenow/sys/sys-account/server/pkg/crud"
	log "github.com/sirupsen/logrus"
)

type Account struct {
	ID                string
	Name              string
	Email             string
	Password          string
	RoleId            string
	UserDefinedFields map[string]interface{}
	CreatedAt         int64
	UpdatedAt         int64
	LastLogin         int64
	Disabled          bool
}

func (a Account) TableName() string {
	return tableName(AccTableName, "_")
}

func accountToQueryParams(acc *Account) (res QueryParams, err error) {
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
func (a Account) CreateSQL() []string {
	fields := initFields(AccColumns)
	tbl := crud.NewTable(AccTableName, fields)
	return []string{tbl.CreateTable()}
}

func (a *AccountDB) getAccountSelectStatement(aqp *QueryParams) (string, []interface{}, error) {
	baseStmt := sq.Select(AccColumns).From(AccTableName)
	if aqp != nil && aqp.Params != nil {
		for k, v := range aqp.Params {
			baseStmt = baseStmt.Where(sq.Eq{k: v})
		}
	}
	return baseStmt.ToSql()
}

func (a *AccountDB) GetAccount(aqp *QueryParams) (*Account, error) {
	var acc Account
	selectStmt, args, err := a.getAccountSelectStatement(aqp)
	if err != nil {
		return nil, err
	}
	doc, err := a.QueryOne(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = document.StructScan(doc, &acc)
	return &acc, err
}

func (a *AccountDB) ListAccount(aqp *QueryParams) ([]*Account, error) {
	var accs []*Account
	selectStmt, args, err := a.getAccountSelectStatement(aqp)
	if err != nil {
		return nil, err
	}
	res, err := a.Query(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = res.Iterate(func(d document.Document) error {
		var acc Account
		if err = document.StructScan(d, &acc); err != nil {
			return err
		}
		accs = append(accs, &acc)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return accs, nil
}

func (a *AccountDB) InsertAccount(acc *Account) error {
	var allVals [][]interface{}
	aqp, err := accountToQueryParams(acc)
	if err != nil {
		return err
	}
	log.Printf("query params: %v", aqp)
	columns, values := aqp.ColumnsAndValues()
	stmt, args, err := sq.Insert(AccTableName).
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

func (a *AccountDB) UpdateAccount(acc *Account) error {
	aqp, err := accountToQueryParams(acc)
	if err != nil {
		return err
	}
	var values [][]interface{}
	stmt, args, err := sq.Update(AccTableName).SetMap(aqp.Params).ToSql()
	if err != nil {
		return err
	}
	values = append(values, args)
	return a.Exec([]string{stmt}, values)
}

func (a *AccountDB) DeleteAccount(id string) error {
	var values [][]interface{}
	stmt, args, err := sq.Delete(AccTableName).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	values = append(values, args)
	return a.Exec([]string{stmt}, values)
}
