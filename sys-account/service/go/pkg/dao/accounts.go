package dao

import (
	"encoding/json"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/pass"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	log "github.com/sirupsen/logrus"

	//"github.com/getcouragenow/sys/main/pkg"
	// FIX IS:
	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/crud"
)

type Account struct {
	ID                string                 `genji:"id"`
	Email             string                 `genji:"email"`
	Password          string                 `genji:"password"`
	RoleId            string                 `genji:"role_id"`
	UserDefinedFields map[string]interface{} `genji:"user_defined_fields"`
	CreatedAt         int64                  `genji:"created_at"`
	UpdatedAt         int64                  `genji:"updated_at"`
	LastLogin         int64                  `genji:"last_login"`
	Disabled          bool                   `genji:"disabled"`
}

func (a *AccountDB) FromPkgAccount(account *pkg.Account) (*Account, error) {
	role, err := a.FromPkgRole(account.Role, account.Id)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:                account.Id,
		Email:             account.Email,
		Password:          account.Password,
		UserDefinedFields: account.Fields.Fields,
		CreatedAt:         account.CreatedAt,
		UpdatedAt:         account.UpdatedAt,
		LastLogin:         account.LastLogin,
		Disabled:          account.Disabled,
		RoleId:            role.ID,
	}, nil
}

func (a *Account) ToPkgAccount(role *pkg.UserRoles) (*pkg.Account, error) {
	createdAt := time.Unix(a.CreatedAt, 0)
	updatedAt := time.Unix(a.UpdatedAt, 0)
	lastLogin := time.Unix(a.LastLogin, 0)
	return &pkg.Account{
		Id:        a.ID,
		Email:     a.Email,
		Password:  a.Password,
		Role:      role,
		CreatedAt: createdAt.Unix(),
		UpdatedAt: updatedAt.Unix(),
		LastLogin: lastLogin.Unix(),
		Disabled:  a.Disabled,
		Fields:    &pkg.UserDefinedFields{Fields: a.UserDefinedFields},
	}, nil
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
	fields := initFields(AccColumns, AccColumnsType)
	tbl := crud.NewTable(AccTableName, fields)
	return []string{
		tbl.CreateTable(),
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email)",
	}
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
	passwd, err := pass.GenHash(acc.Password)
	if err != nil {
		return err
	}
	acc.Password = passwd
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
	}).Info("INSERT to accounts table")
	if err != nil {
		return err
	}
	return a.Exec(stmt, values)
}

func (a *AccountDB) UpdateAccount(acc *Account) error {
	aqp, err := accountToQueryParams(acc)
	if err != nil {
		return err
	}
	stmt, args, err := sq.Update(AccTableName).SetMap(aqp.Params).ToSql()
	if err != nil {
		return err
	}
	return a.Exec(stmt, args)
}

func (a *AccountDB) DeleteAccount(id string) error {
	stmt, args, err := sq.Delete(AccTableName).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	return a.Exec(stmt, args)
}
