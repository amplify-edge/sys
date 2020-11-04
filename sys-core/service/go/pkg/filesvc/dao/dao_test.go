package dao_test

import (
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	corecfg "github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc/dao"
	log "github.com/sirupsen/logrus"
	"testing"
)

var (
	fdb    *dao.FileDB
	err    error

	project1ID = coredb.NewID()
	org1ID     = coredb.NewID()
	account1ID = coredb.NewID()
)

func init() {
	var csc *corecfg.FileServiceConfig
	csc, err = corecfg.NewConfig("./testdata/db.yml")
	if err != nil {
		log.Fatalf("error initializing db: %v", err)
	}
	logger := log.New().WithField("test", "sys-file")
	logger.Level = log.DebugLevel
	testDb, err := coredb.NewCoreDB(logger, &csc.DBConfig, nil)
	if err != nil {
		log.Fatalf("error creating CoreDB: %v", err)
	}
	log.Debug("MakeSchema testing .....")
	fdb, err = dao.NewFileDB(testDb, logger)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully initialize sys-file-db:  %v", fdb)
}

func TestAll(t *testing.T) {
	t.Run("Test Upsert File", testUpsertFile)
}

func testUpsertFile(t *testing.T) {
}
