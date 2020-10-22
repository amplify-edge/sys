package coredb

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	backupFormat = "%s_%s.bak"
)

func (c *CoreDB) scheduleBackup() error {
	crony := cron.New()
	errChan := make(chan error, 1)
	_, err := crony.AddFunc(c.config.SysCoreConfig.CronConfig.BackupSchedule, func() {
		c.logger.Debug("creating backup schedule")
		fileWriter, err := c.createBackupFile()
		defer fileWriter.Close()
		if err != nil {
			c.logger.Debugf("%s error while creating backup file: %v", moduleName, err)
			errChan <- err
			return
		}
		badgerDb := c.engine.DB
		// full backup, no matter what
		// TODO: provide incremental backup as well perhaps?
		_, err = badgerDb.Backup(fileWriter, 0)
		if err != nil {
			c.logger.Debugf("%s error while doing streaming backup: %v", moduleName, err)
			errChan <- err
			return
		}
	})
	close(errChan)
	if errFromChan := <-errChan; errFromChan != nil {
		return errFromChan
	}
	if err != nil {
		return err
	}

	// TODO: rotate encryption key
	// Find a way to do streaming backup while re-encrypting the key perhaps?
	c.crony = crony
	return nil
}

func (c *CoreDB) createBackupFile() (io.WriteCloser, error) {
	currentTime := time.Now().Format("200601021859")
	backupFileName := filepath.Join(
		c.config.SysCoreConfig.CronConfig.BackupDir,
		fmt.Sprintf(backupFormat, c.config.SysCoreConfig.DbConfig.Name, currentTime),
	)
	return createFile(backupFileName)
}

func createFile(fileName string) (io.WriteCloser, error) {
	f, err := os.Create(fileName)
	return f, err
}
