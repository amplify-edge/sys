package coredb

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji"
)

func (c *CoreDB) Query(stmt string, args ...interface{}) (*QueryResult, error) {
	res, err := c.store.Query(stmt, args...)
	if err != nil {
		return nil, err
	}
	return &QueryResult{
		res,
	}, nil
}

func (c *CoreDB) QueryOne(stmt string, args ...interface{}) (*DocumentResult, error) {
	res, err := c.store.QueryDocument(stmt, args...)
	if err != nil {
		return nil, err
	}
	return &DocumentResult{
		res,
	}, nil
}

func (c *CoreDB) createTx() (*genji.Tx, error) {
	return c.store.Begin(true)
}

func (c *CoreDB) txhelper(tx *genji.Tx, operation func() error) error {
	defer tx.Rollback()
	if err := operation(); err != nil {
		return err
	}
	return tx.Commit()
}

func (c *CoreDB) Exec(stmt string, args ...interface{}) error {
	tx, err := c.createTx()
	if err != nil {
		return err
	}
	return c.txhelper(tx, func() error {
		return tx.Exec(stmt, args...)
	})
}

func (c *CoreDB) BulkExec(stmtMap map[string][]interface{}) error {
	tx, err := c.createTx()
	if err != nil {
		return err
	}
	return c.txhelper(tx, func() error {
		for k, v := range stmtMap {
			if err := tx.Exec(k, v...); err != nil {
				return err
			}
		}
		return nil
	})
}

type StmtIFacer interface {
	ToSql() (string, []interface{}, error)
}

func BaseQueryBuilder(filter map[string]interface{}, tableName, tableColumns string, stmtBuilderFunc func(k string, v interface{}) StmtIFacer) sq.SelectBuilder {
	baseStmt := sq.Select(tableColumns).From(tableName)
	if filter != nil {
		for k, v := range filter {
			baseStmt = baseStmt.Where(stmtBuilderFunc(k, v))
		}
	}
	return baseStmt
}

func ListSelectStatement(baseStmt sq.SelectBuilder, orderBy string, limit int64, cursor *int64, cursorName string) (string, []interface{}, error) {
	csr := *cursor
	if cursor == nil {
		csr = 0
	}
	baseStmt = baseStmt.Where(sq.Gt{cursorName: csr})
	baseStmt = baseStmt.Limit(uint64(limit)).OrderBy(orderBy)
	return baseStmt.ToSql()
}
