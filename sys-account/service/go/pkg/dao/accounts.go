package dao

import (
	"encoding/json"
	"fmt"
	"github.com/genjidb/genji/document"
	"time"

	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/pass"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

var (
	accountsUniqueIdx = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_email ON %s(email)", AccTableName)
)

type Account struct {
	ID                string                 `genji:"id"`
	Email             string                 `genji:"email"`
	Password          string                 `genji:"password"`
	RoleId            string                 `genji:"role_id"`
	UserDefinedFields map[string]interface{} `genji:"user_defined_fields"`
	Survey            map[string]interface{} `genji:"survey"`
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
		Survey:            account.Survey.Fields,
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
		Survey:    &pkg.UserDefinedFields{Fields: a.Survey},
	}, nil
}

func accountToQueryParams(acc *Account) (res coresvc.QueryParams, err error) {
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
func (a Account) CreateSQL() []string {
	fields := initFields(AccColumns, AccColumnsType)
	tbl := coresvc.NewTable(AccTableName, fields, []string{accountsUniqueIdx})
	return tbl.CreateTable()
}

func (a *AccountDB) accountQueryFilter(filter *coresvc.QueryParams) sq.SelectBuilder {
	baseStmt := sq.Select(AccColumns).From(AccTableName)
	if filter != nil && filter.Params != nil {
		for k, v := range filter.Params {
			baseStmt = baseStmt.Where(sq.Eq{k: v})
		}
	}
	return baseStmt
}

func (a *AccountDB) GetAccount(filterParams *coresvc.QueryParams) (*Account, error) {
	var acc Account
	selectStmt, args, err := a.accountQueryFilter(filterParams).ToSql()
	if err != nil {
		return nil, err
	}
	doc, err := a.db.QueryOne(selectStmt, args...)
	if err != nil {
		return nil, err
	}
	err = doc.StructScan(&acc)
	return &acc, err
}

func (a *AccountDB) ListAccount(filterParams *coresvc.QueryParams, orderBy string, limit, cursor int64) ([]*Account, int64, error) {
	var accs []*Account
	baseStmt := a.accountQueryFilter(filterParams)
	selectStmt, args, err := a.listSelectStatements(baseStmt, orderBy, limit, &cursor)
	if err != nil {
		return nil, 0, err
	}
	res, err := a.db.Query(selectStmt, args...)
	if err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}
	return accs, accs[len(accs)-1].CreatedAt, nil
}

func (a *AccountDB) InsertAccount(acc *Account) error {
	passwd, err := pass.GenHash(acc.Password)
	if err != nil {
		return err
	}
	acc.Password = passwd
	filterParams, err := accountToQueryParams(acc)
	if err != nil {
		return err
	}
	log.Debugf("query params: %v", filterParams)
	columns, values := filterParams.ColumnsAndValues()
	stmt, args, err := sq.Insert(AccTableName).
		Columns(columns...).
		Values(values...).
		ToSql()
	a.log.WithFields(log.Fields{
		"statement": stmt,
		"args":      args,
	}).Debug("insert to accounts table")
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) UpdateAccount(acc *Account) error {
	filterParams, err := accountToQueryParams(acc)
	if err != nil {
		return err
	}
	stmt, args, err := sq.Update(AccTableName).SetMap(filterParams.Params).ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}

func (a *AccountDB) DeleteAccount(id string) error {
	stmt, args, err := sq.Delete(AccTableName).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	return a.db.Exec(stmt, args...)
}
