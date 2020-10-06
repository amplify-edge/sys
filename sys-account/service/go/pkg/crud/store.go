// package crud should be in sys-core i think
package crud

import (
	"bytes"
	"text/template"

	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/sql/query"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/utilities"
)

const (
	CreateTableTpl = `CREATE TABLE IF NOT EXISTS {{ .Name | toSnakeCase }} (
		{{ $s := separator ",\n" }}{{ range $k, $v := .Fields }}{{ call $s}}{{ $k | toSnakeCase }} {{ $v }}{{ end }} 
	);
`
)

type Table struct {
	Name   string
	Fields map[string]string
}

// DbTxer provides common interface for executing sql statements
type DbTxer interface {
	Exec(string, []interface{}) error
	Query(string, ...interface{}) (query.Result, error)
	QueryOne(string, ...interface{}) (document.Document, error)
	BuildSearchQuery(string) string
}

func NewTable(name string, fields map[string]string) *Table {
	return &Table{name, fields}
}

func (t *Table) CreateTable() string {
	funcMap := template.FuncMap{
		"separator":   utilities.SeparatorFunc,
		"toSnakeCase": utilities.ToSnakeCase,
	}
	tpl := template.Must(template.New("createTable").
		Funcs(funcMap).Parse(CreateTableTpl))
	var bf bytes.Buffer
	tpl.Execute(&bf, t)

	return bf.String()
}
