package service

import (
	"bytes"
	"github.com/dgraph-io/badger/v2"
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/engine/badgerengine"
	"github.com/genjidb/genji/sql/query"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"text/template"
	"time"

	corecfg "github.com/getcouragenow/sys/sys-core/service/go"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/internal/helper"
)

const (
	day = 24 * time.Hour
)

// CoreDB is the exported struct
type CoreDB struct {
	logger *log.Entry
	store  *genji.DB
	engine *badgerengine.Engine
	models map[string]DbModel
	config *corecfg.SysCoreConfig
	crony  *cron.Cron
}

// NewCoreDB facilitates creation of (wrapped) genji database alongside badger DB engine
// if one wants to use one or the other.
// or if internally will use the underlying badger DB engine to create Stream for example
// for backup, restore, or anything
func NewCoreDB(l *log.Entry, cfg *corecfg.SysCoreConfig) (*CoreDB, error) {
	dbName := cfg.SysCoreConfig.DbConfig.Name
	dbPath := cfg.SysCoreConfig.DbConfig.DbDir + "/" + dbName
	store, engine, err := newGenjiStore(dbPath, cfg.SysCoreConfig.DbConfig.EncryptKey, cfg.SysCoreConfig.DbConfig.RotationDuration)
	if err != nil {
		return nil, err
	}
	return &CoreDB{
		logger: l,
		store:  store,
		engine: engine,
		models: map[string]DbModel{},
		config: cfg,
	}, nil
}

// helper function to create genji.DB
func newGenjiStore(path string, encKey string, keyRotationSchedule int) (*genji.DB, *badgerengine.Engine, error) {
	// badgerengine options with encryption and encryption key rotation
	options := badger.DefaultOptions(path).
		WithEncryptionKey(helper.MD5(encKey)) // .
	// TODO: encryption key rotation is currently disabled, which is not great
	// WithEncryptionKeyRotationDuration(time.Duration(keyRotationSchedule) * day)
	engine, err := badgerengine.NewEngine(options)
	if err != nil {
		return nil, nil, err
	}
	store, err := genji.New(engine)
	if err != nil {
		return nil, nil, err
	}
	return store, engine, nil
}

const (
	createTableTpl = `CREATE TABLE IF NOT EXISTS {{ .Name | toSnakeCase }} (
		{{ $s := separator ",\n" }}{{ range $k, $v := .Fields }}{{ call $s}}{{ $k | toSnakeCase }} {{ $v }}{{ end }} 
	);
`
)

type Table struct {
	Name            string
	Fields          map[string]string
	IndexStatements []string
}

func NewTable(name string, fields map[string]string, indexStatements []string) *Table {
	return &Table{name, fields, indexStatements}
}

// Utility function for each consumer to create their own module
// each module will only then have to call this function to satisfy
// DBModel interface below
func (t *Table) CreateTable() []string {
	var tblInitStatements []string
	funcMap := template.FuncMap{
		"separator":   helper.SeparatorFunc,
		"toSnakeCase": helper.ToSnakeCase,
	}
	tpl := template.Must(template.New("createTable").
		Funcs(funcMap).Parse(createTableTpl))
	var bf bytes.Buffer
	tpl.Execute(&bf, t)
	tblInitStatements = append(tblInitStatements, bf.String())
	tblInitStatements = append(tblInitStatements, t.IndexStatements...)
	return tblInitStatements
}

// DbModel Basic table model interface,
type DbModel interface {
	CreateSQL() []string
}

type QueryResult struct {
	*query.Result
}

type DocumentResult struct {
	document.Document
}

func (d *DocumentResult) StructScan(dest interface{}) error {
	return d.StructScan(dest)
}
