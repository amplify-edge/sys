package db

import (
	"fmt"
	"os"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/engine/badgerengine"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"

	service "github.com/getcouragenow/sys/sys-core/service/go"
)

var (
	database *genji.DB = nil
	models             = make(map[string][]DbModel)
	config   *service.SysCoreConfig
	bakcron  *BackupCron
)

// UID Generate ksuid.
func UID() string {
	return ksuid.New().String()
}

// DbModel Basic table model interface,
type DbModel interface {
	// Table name
	TableName() string
	// Used to return the SQL used to create tables and indexes
	CreateSQL() []string
}

// InitDatabase Must be initialized before using database.
func InitDatabase(cfg *service.SysCoreConfig) error {
	config = cfg

	if exists, _ := PathExists(cfg.DbConfig.DbDir); !exists {
		_ = os.MkdirAll(cfg.DbConfig.DbDir, os.ModePerm)
	}

	if exists, _ := PathExists(cfg.CronConfig.BackupDir); !exists {
		_ = os.MkdirAll(cfg.CronConfig.BackupDir, os.ModePerm)
	}

	if database != nil {
		return fmt.Errorf("The database has been initialized")
	}

	dbName := config.DbConfig.Name
	dbPath := cfg.DbConfig.DbDir + "/" + dbName
	log.Print("Db " + dbPath + "Open .....")
	var err error
	database, err = makeDb(dbPath, config.DbConfig.EncryptKey)
	if err != nil {
		log.Fatalf("Db "+dbPath+"Open failed: %v", err)
		return err
	}
	bakcron = NewBackupCron(cfg)
	// TODO: Causes concurrency issues. Needs rethink..
	// bakcron.Start()
	return nil
}

// SharedDatabase Returns the global Genji database shared pointer.
func SharedDatabase() (*genji.DB, error) {
	if database == nil {
		err := fmt.Errorf("DB is not initialized, please call InitDatabase first")
		log.Fatal(err)
		return nil, err
	}
	return database, nil
}

func makeDb(name string, key string) (*genji.DB, error) {
	// Create a badger engine
	options := badger.DefaultOptions(name)
	if len(key) == 0 {
		return nil, fmt.Errorf("[%s] Invalid encryption key", name)
	}
	// The key length must be 16 or 32, so use md5 to encrypt once.
	options.EncryptionKey = MD5(key)
	// Set to 180 days or something else?
	options.EncryptionKeyRotationDuration = 180 * 24 * time.Hour
	ng, err := badgerengine.NewEngine(badger.DefaultOptions(name))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var db *genji.DB
	// Pass it to genji
	db, err = genji.New(ng)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return db, nil
}

// RegisterModels for mod-*, For example, mod-accounts, mod-chat, sys-core.
func RegisterModels(mod string, mds []DbModel) {
	models[mod] = mds
}

// MakeSchema create tables for `accounts` if not exists.
func MakeSchema(gdb *genji.DB) error {
	// DO in a transaction
	tx, err := gdb.Begin(true)
	if err != nil {
		log.Panic(err)
		return err
	}
	defer tx.Rollback()

	for k, tables := range models {
		log.Printf("MakeSchema for: %v", k)
		for _, table := range tables {
			log.Printf("Create Table: %v, sql = %v", table.TableName(), table.CreateSQL()[0])
			for _, sql := range table.CreateSQL() {
				if err := gdb.Exec(sql); err != nil {
					log.Error(err)
					return err
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func QueryTable(db *genji.DB, in interface{}, sql string, outcb func(out interface{})) error {
	stream, err := db.Query(sql)
	if err != nil {
		log.Print(err)
		return err
	}

	defer stream.Close()

	// Iterate over the results
	var out interface{} = in
	err = stream.Iterate(func(d document.Document) error {
		err = document.StructScan(d, out)
		if err != nil {
			return err
		}
		// log.Printf("out => %v", out)
		outcb(out)
		return nil
	})

	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}
