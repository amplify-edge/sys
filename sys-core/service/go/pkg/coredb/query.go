package coredb

import (
	"database/sql/driver"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji"
	"reflect"
	"sort"
	"strings"
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

func buildSearchQuery(qs string) string {
	var sb strings.Builder
	sb.WriteString("%")
	sb.WriteString(qs)
	sb.WriteString("%")
	return sb.String()
}

const (
	// Portable true/false literals.
	sqlTrue  = "(1=1)"
	sqlFalse = "(1=0)"
)

type sqIn map[string]interface{}

type sqNotIn sqIn

func getSortedKeys(exp map[string]interface{}) []string {
	sortedKeys := make([]string, 0, len(exp))
	for k := range exp {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func isListType(val interface{}) bool {
	if driver.IsValue(val) {
		return false
	}
	valVal := reflect.ValueOf(val)
	return valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice
}

func placeholders(count int) string {
	if count < 1 {
		return ""
	}

	return strings.Repeat(",?", count)[1:]
}

func (in sqIn) toSQL(useNotOpr bool) (sql string, args []interface{}, err error) {
	if len(in) == 0 {
		// Empty Sql{} evaluates to true.
		sql = sqlTrue
		return
	}

	var (
		exprs       []string
		equalOpr    = "="
		inOpr       = "IN"
		nullOpr     = "IS"
		inEmptyExpr = sqlFalse
	)

	if useNotOpr {
		equalOpr = "<>"
		inOpr = "NOT IN"
		nullOpr = "IS NOT"
		inEmptyExpr = sqlTrue
	}

	sortedKeys := getSortedKeys(in)
	for _, key := range sortedKeys {
		var expr string
		val := in[key]

		switch v := val.(type) {
		case driver.Valuer:
			if val, err = v.Value(); err != nil {
				return
			}
		}

		r := reflect.ValueOf(val)
		if r.Kind() == reflect.Ptr {
			if r.IsNil() {
				val = nil
			} else {
				val = r.Elem().Interface()
			}
		}

		if val == nil {
			expr = fmt.Sprintf("%s %s NULL", key, nullOpr)
		} else {
			if isListType(val) {
				valVal := reflect.ValueOf(val)
				if valVal.Len() == 0 {
					expr = inEmptyExpr
					if args == nil {
						args = []interface{}{}
					}
				} else {
					for i := 0; i < valVal.Len(); i++ {
						args = append(args, valVal.Index(i).Interface())
					}
					expr = fmt.Sprintf("%s %s [%s]", key, inOpr, placeholders(valVal.Len()))
				}
			} else {
				expr = fmt.Sprintf("%s %s ?", key, equalOpr)
				args = append(args, val)
			}
		}
		exprs = append(exprs, expr)
	}
	sql = strings.Join(exprs, " AND ")
	return
}

func (in sqIn) ToSql() (sql string, args []interface{}, err error) {
	return in.toSQL(false)
}

func (in sqNotIn) ToSql() (sql string, args []interface{}, err error) {
	return sqIn(in).toSQL(true)
}

func matchStmtBuilderFunc(want string) func(k string, v interface{}) StmtIFacer {
	return func(k string, v interface{}) StmtIFacer {
		switch ToSnakeCase(want) {
		case "like":
			return sq.Like{k: buildSearchQuery(v.(string))}
		case "eq":
			return sq.Eq{k: v}
		case "in":
			return sqIn{k: v}
		case "not_in", "not_eq":
			return sq.NotEq{k: v}
		case "gt":
			return sq.Gt{k: v}
		case "gte":
			return sq.GtOrEq{k: v}
		case "lte":
			return sq.LtOrEq{k: v}
		case "lt":
			return sq.Lt{k: v}
		default:
			return sq.Like{k: buildSearchQuery(v.(string))}
		}
	}
}

func BaseQueryBuilder(filter map[string]interface{}, tableName, tableColumns string, sqlMatcher string) sq.SelectBuilder {
	baseStmt := sq.Select(tableColumns).From(tableName)
	if filter != nil {
		for k, v := range filter {
			baseStmt = baseStmt.Where(matchStmtBuilderFunc(sqlMatcher)(k, v))
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
