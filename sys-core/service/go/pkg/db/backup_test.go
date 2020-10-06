package db_test

import (
	"log"
	"os"
	"testing"

	"github.com/genjidb/genji"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
)

var (
	testDb    *genji.DB
	backupDir = "backups"
)

func init() {
	testDb, _ = db.SharedDatabase()
}

func TestBackup(t *testing.T) {
	if exists, _ := db.PathExists(backupDir); !exists {
		os.MkdirAll(backupDir, os.ModePerm)
	}
	backupFile := "./" + backupDir + "/" + "db_backup.bak"
	dbName := "getcouragenow.db"
	log.Printf("backup %v => %v", dbName, backupFile)
	//Db must be closed first
	testDb.Close()
	if err := db.BackupDb(dbName, backupFile); err != nil {
		t.Error(err)
	}
}
