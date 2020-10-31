package coredb

import (
	"bytes"
	"github.com/dgraph-io/badger/v2"
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/engine/badgerengine"
	"github.com/genjidb/genji/sql/query"
	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"
	"github.com/robfig/cron/v3"
	"github.com/segmentio/encoding/json"
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
	logger    *log.Entry
	store     *genji.DB
	engine    *badgerengine.Engine
	models    map[string]DbModel
	config    *corecfg.SysCoreConfig
	crony     *cron.Cron
	cronFuncs map[string]func()
}

// NewCoreDB facilitates creation of (wrapped) genji database alongside badger DB engine
// if one wants to use one or the other.
// or if internally will use the underlying badger DB engine to create Stream for example
// for backup, restore, or anything
func NewCoreDB(l *log.Entry, cfg *corecfg.SysCoreConfig, cronFuncs map[string]func()) (*CoreDB, error) {
	dbName := cfg.SysCoreConfig.DbConfig.Name
	dbPath := cfg.SysCoreConfig.DbConfig.DbDir + "/" + dbName
	store, engine, err := newGenjiStore(dbPath, cfg.SysCoreConfig.DbConfig.EncryptKey, cfg.SysCoreConfig.DbConfig.RotationDuration)
	if err != nil {
		return nil, err
	}
	cdb := &CoreDB{
		logger:    l,
		store:     store,
		engine:    engine,
		models:    map[string]DbModel{},
		config:    cfg,
		cronFuncs: cronFuncs,
	}
	err = cdb.scheduleBackup()
	if err != nil {
		return nil, err
	}
	cdb.crony.Start()
	return cdb, nil
}

// helper function to create genji.DB
func newGenjiStore(path, encKey string, keyRotationSchedule int) (*genji.DB, *badgerengine.Engine, error) {
	// badgerengine options with encryption and encryption key rotation
	options := createBadgerOpts(path, encKey, keyRotationSchedule)
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

func createBadgerOpts(path, encKey string, keyRotationSchedule int) badger.Options {
	return badger.DefaultOptions(path).
		WithEncryptionKey(helper.MD5(encKey))
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
	if err := tpl.Execute(&bf, t); err != nil {
		log.Fatal(err)
	}
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
	Doc document.Document
}

func (d *DocumentResult) StructScan(dest interface{}) error {
	return document.StructScan(d.Doc, dest)
}

type QueryParams struct {
	Params map[string]interface{}
}

func (qp *QueryParams) ColumnsAndValues() ([]string, []interface{}) {
	var columns []string
	var values []interface{}
	for k, v := range qp.Params {
		columns = append(columns, helper.ToSnakeCase(k))
		values = append(values, v)
	}
	return columns, values
}

func UnmarshalToMap(b []byte) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func MarshalToBytes(any interface{}) ([]byte, error) {
	return json.Marshal(&any)
}

func AnyToQueryParam(m interface{}, snakeCase bool) (res QueryParams, err error) {
	jbytes, err := MarshalToBytes(&m)
	if err != nil {
		return QueryParams{}, err
	}
	params, err := UnmarshalToMap(jbytes)
	if err != nil {
		return QueryParams{}, err
	}
	if snakeCase {
		for k, v := range params {
			key := ToSnakeCase(k)
			val := v
			delete(params, k)
			params[key] = val
		}
	}
	res.Params = params
	return res, err
}

func NewID() string {
	return sharedConfig.NewID()
}

func CurrentTimestamp() int64 {
	return time.Now().UTC().Unix()
}

func ToSnakeCase(s string) string {
	return helper.ToSnakeCase(s)
}
