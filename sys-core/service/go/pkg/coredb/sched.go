package coredb

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/types/known/emptypb"

	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"
	sharedPkg "github.com/getcouragenow/sys-share/sys-core/service/go/pkg"
)

const (
	backupFormat = "%s_%s.bak"
)

func (c *CoreDB) RegisterCronFunction(funcSpec string, function func()) error {
	_, err := c.crony.AddFunc(funcSpec, function)
	if err != nil {
		return err
	}
	return nil
}

func (c *CoreDB) scheduleBackup() error {
	crony := cron.New()

	// default backup schedule
	errChan := make(chan error, 1)
	_, err := crony.AddFunc(c.config.CronConfig.BackupSchedule, func() {
		_, err := c.backup()
		if err != nil {
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

	// custom cron functions from each module
	if c.cronFuncs != nil && len(c.cronFuncs) > 0 {
		for funcSpec, fun := range c.cronFuncs {
			errChan := make(chan error, 1)
			_, err := crony.AddFunc(funcSpec, fun)
			close(errChan)
			if errFromChan := <-errChan; errFromChan != nil {
				return errFromChan
			}
			if err != nil {
				return err
			}
		}
	}
	// TODO: rotate encryption key
	// Find a way to do streaming backup while re-encrypting the key perhaps?
	c.crony = crony
	return nil
}

func (c *CoreDB) Backup(ctx context.Context, in *emptypb.Empty) (*sharedPkg.BackupResult, error) {
	filename, err := c.backup()
	if err != nil {
		return nil, err
	}
	return &sharedPkg.BackupResult{BackupFile: filename}, nil
}

func (c *CoreDB) backup() (string, error) {
	c.logger.Debug("creating backup schedule")
	fileWriter, filename, err := c.createBackupFile()
	defer fileWriter.Close()
	if err != nil {
		c.logger.Debugf("%s error while creating backup file: %v", moduleName, err)
		return "", err
	}
	badgerDb := c.engine.DB
	// full backup, no matter what
	// TODO: provide incremental backup as well perhaps?
	_, err = badgerDb.Backup(fileWriter, 0)
	if err != nil {
		c.logger.Debugf("%s error while doing streaming backup: %v", moduleName, err)
		return "", err
	}
	return filename, nil
}

func (c *CoreDB) Restore(_ context.Context, in *sharedPkg.RestoreRequest) (*sharedPkg.RestoreResult, error) {
	badgerDB := c.engine.DB
	f, err := c.openFile(in.BackupFile)
	if err != nil {
		return nil, err
	}
	err = badgerDB.Load(f, 10)
	if err != nil {
		return nil, err
	}
	return &sharedPkg.RestoreResult{Result: fmt.Sprintf("successfully restore db: %s", in.BackupFile)}, nil
}

func (c *CoreDB) ListBackup(ctx context.Context, in *emptypb.Empty) (*sharedPkg.ListBackupResult, error) {
	var bfiles []*sharedPkg.BackupResult
	listFiles, err := c.listBackups()
	if err != nil {
		return nil, err
	}
	for _, f := range listFiles {
		bfiles = append(bfiles, &sharedPkg.BackupResult{BackupFile: f})
	}
	return &sharedPkg.ListBackupResult{BackupFiles: bfiles}, nil
}

func (c *CoreDB) listBackups() ([]string, error) {
	backupDir := c.config.CronConfig.BackupDir
	c.logger.Info("backup dir: " + backupDir)
	fileInfos, err := sharedConfig.ListFiles(backupDir)
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, f := range fileInfos {
		filenames = append(filenames, f.Name())
	}
	return filenames, nil
}

func (c *CoreDB) createBackupFile() (io.WriteCloser, string, error) {
	currentTime := time.Now().Format("200601021859")
	backupFileName := filepath.Join(
		c.config.CronConfig.BackupDir,
		fmt.Sprintf(backupFormat, c.config.DbConfig.Name, currentTime),
	)
	f, err := createFile(backupFileName)
	if err != nil {
		return nil, "", err
	}
	return f, backupFileName, nil
}

func createFile(fileName string) (io.WriteCloser, error) {
	f, err := os.Create(fileName)
	return f, err
}

func (c *CoreDB) openFile(filepath string) (io.ReadCloser, error) {
	exists := sharedConfig.FileExists(filepath)
	if !exists {
		return nil, fmt.Errorf("cannot find %s", filepath)
	}
	return os.Open(filepath)
}
