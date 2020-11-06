package dao

import (
	"fmt"
	"github.com/genjidb/genji/document"
	"time"

	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"

	utilities "github.com/getcouragenow/sys-share/sys-core/service/config"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/pass"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

var (
	accountsUniqueIdx      = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_email ON %s(email)", AccTableName, AccTableName)
	accountAvatarUniqueIdx = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_avatar_resource_id ON %s(avatar_resource_id)", AccTableName, AccTableName)
)

type Account struct {
	ID                string `json:"id,omitempty" genji:"id" coredb:"primary"`
	Email             string `json:"email,omitempty" genji:"email"`
	Password          string `json:"password,omitempty" genji:"password"`
	CreatedAt         int64  `json:"created_at" genji:"created_at"`
	UpdatedAt         int64  `json:"updated_at" genji:"updated_at"`
	LastLogin         int64  `json:"last_login" genji:"last_login"`
	Disabled          bool   `json:"disabled" genji:"disabled"`
	Verified          bool   `json:"verified" genji:"verified"`
	VerificationToken string `json:"verification_token,omitempty" genji:"verification_token"`
	AvatarResourceId  string `json:"avatar_resource_id,omitempty" genji:"avatar_resource_id"`
}

func (a *AccountDB) InsertFromPkgAccountRequest(account *pkg.AccountNewRequest, verified bool) (*Account, error) {
	accountId := utilities.NewID()
	var roles []*Role
	if account.Roles != nil && len(account.Roles) > 0 {
		a.log.Debugf("Convert and getting roles")
		for _, pkgRole := range account.Roles {
			role := a.FromPkgRoleRequest(pkgRole, accountId)
			roles = append(roles, role)
		}
	} else {
		roles = append(roles, &Role{
			ID:        coresvc.NewID(),
			AccountId: accountId,
			Role:      int(pkg.GUEST),
			ProjectId: "",
			OrgId:     "",
			CreatedAt: time.Now().UTC().Unix(),
		})
	}
	for _, daoRole := range roles {
		err := a.InsertRole(daoRole)
		if err != nil {
			return nil, err
		}
	}
	isVerified := false
	if verified {
		isVerified = verified
	}
	acc := &Account{
		ID:               accountId,
		Email:            account.Email,
		Password:         account.Password,
		CreatedAt:        coresvc.CurrentTimestamp(),
		UpdatedAt:        coresvc.CurrentTimestamp(),
		LastLogin:        coresvc.CurrentTimestamp(),
		Disabled:         false,
		Verified:         isVerified,
		AvatarResourceId: account.AvatarFilepath,
	}

	if err := a.InsertAccount(acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (a *AccountDB) FromPkgAccount(account *pkg.Account) (*Account, error) {
	return &Account{
		ID:               account.Id,
		Email:            account.Email,
		Password:         account.Password,
		CreatedAt:        account.CreatedAt,
		UpdatedAt:        account.UpdatedAt,
		LastLogin:        account.LastLogin,
		Disabled:         account.Disabled,
		Verified:         account.Verified,
		AvatarResourceId: account.AvatarResourceId,
	}, nil
}

func (a *Account) ToPkgAccount(roles []*pkg.UserRoles, avatar []byte) (*pkg.Account, error) {
	createdAt := time.Unix(a.CreatedAt, 0)
	updatedAt := time.Unix(a.UpdatedAt, 0)
	lastLogin := time.Unix(a.LastLogin, 0)
	return &pkg.Account{
		Id:               a.ID,
		Email:            a.Email,
		Password:         a.Password,
		Role:             roles,
		CreatedAt:        createdAt.Unix(),
		UpdatedAt:        updatedAt.Unix(),
		LastLogin:        lastLogin.Unix(),
		Disabled:         a.Disabled,
		Verified:         a.Verified,
		AvatarResourceId: a.AvatarResourceId,
		Avatar:           avatar,
	}, nil
}

func accountToQueryParams(acc *Account) (res coresvc.QueryParams, err error) {
	return coresvc.AnyToQueryParam(acc, true)
}

// CreateSQL will only be called once by sys-core see sys-core API.
func (a Account) CreateSQL() []string {
	fields := coresvc.GetStructTags(a)
	tbl := coresvc.NewTable(AccTableName, fields, []string{accountsUniqueIdx, accountAvatarUniqueIdx})
	return tbl.CreateTable()
}

func (a *AccountDB) accountQueryFilter(filter *coresvc.QueryParams) sq.SelectBuilder {
	baseStmt := sq.Select(a.accountColumns).From(AccTableName)
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
	res.Close()
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
	delete(filterParams.Params, "id")
	stmt, args, err := sq.Update(AccTableName).SetMap(filterParams.Params).
		Where(sq.Eq{"id": acc.ID}).ToSql()
	a.log.Debugf(
		"update accounts statement: %v, args: %v", stmt,
		args,
	)
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
	rstmt, rargs, err := sq.Delete(RolesTableName).Where("account_id = ?", id).ToSql()
	if err != nil {
		return err
	}
	return a.db.BulkExec(map[string][]interface{}{
		stmt:  args,
		rstmt: rargs,
	})
}
