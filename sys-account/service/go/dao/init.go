package dao

import (
	"errors"
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/sql/query"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
	"github.com/sirupsen/logrus"
	"strings"
)

var (
	tablePrefix = "sys_core"
	modName     = "accounts"
)

type AccountDB struct {
	db  *genji.DB
	log *logrus.Logger
}

func NewAccountDB(db *genji.DB) (*AccountDB, error) {
	tables := []coredb.DbModel{
		Account{},
		Permission{},
	}
	coredb.RegisterModels(modName, tables)
	if err := coredb.MakeSchema(db); err != nil {
		return nil, err
	}
	log := logrus.New()
	return &AccountDB{db, log}, nil
}

func (a *AccountDB) Exec(stmts []string, argSlices [][]interface{}) error {
	if len(stmts) != len(argSlices) {
		return errors.New("mismatch statements and argument counts")
	}
	tx, err := a.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, stmt := range stmts {
		if err := tx.Exec(stmt, argSlices[i]...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (a *AccountDB) Query(stmt string, args ...interface{}) (*query.Result, error) {
	return a.db.Query(stmt, args...)
}

func (a *AccountDB) QueryOne(stmt string, args ...interface{}) (document.Document, error) {
	return a.db.QueryDocument(stmt, args...)
}

func (a *AccountDB) BuildSearchQuery(qs string) string {
	var sb strings.Builder
	sb.WriteString("%")
	sb.WriteString(qs)
	sb.WriteString("%")
	return sb.String()
}
