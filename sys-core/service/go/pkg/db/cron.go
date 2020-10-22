package db

import (
	"fmt"
	"log"
	"time"

	service "github.com/getcouragenow/sys/sys-core/service/go"
	"github.com/robfig/cron/v3"
)

type BackupCron struct {
	cron       *cron.Cron
	backupSpec string
	rotateSpec string
}

func NewBackupCron(config *service.SysCoreConfig) *BackupCron {
	return &BackupCron{
		cron:       cron.New(),
		backupSpec: config.SysCoreConfig.CronConfig.BackupSchedule,
		rotateSpec: config.SysCoreConfig.CronConfig.RotateSchedule,
	}
}

func (bc *BackupCron) Start() {
	bc.cron.AddFunc(bc.backupSpec, func() {
		fmt.Println("Do backup schedule!")
		currentTime := time.Now().Format("200601021859")
		backupFile := config.SysCoreConfig.CronConfig.BackupDir + "/" + "db_" + currentTime + ".bak"
		dbPath := config.SysCoreConfig.DbConfig.DbDir + "/" + config.SysCoreConfig.DbConfig.Name
		// Must be closed before backup

		// TODO: need a sync.Locker
		database.Close()
		BackupDb(dbPath, backupFile)
		// Reopen the data
		var err error
		database, err = makeDb(dbPath, config.SysCoreConfig.DbConfig.EncryptKey)
		if err != nil {
			log.Fatalf("Db "+dbPath+"Open failed: %v", err)
		}
		// TODO: Send notification to mod-disco?
	})

	bc.cron.AddFunc(bc.backupSpec, func() {
		fmt.Println("Do db rotate schedule!")
		// TODO: rotate encryption key.
		// - create new key
		// - rotate the key
		// - write a log?
	})
	bc.cron.Start()
}

func (bc *BackupCron) Stop() {
	bc.cron.Stop()
}
