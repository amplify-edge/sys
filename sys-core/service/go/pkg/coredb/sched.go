package coredb

import (
	"context"
	"fmt"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/fileutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/types/known/emptypb"

	sharedConfig "go.amplifyedge.org/sys-share-v2/sys-core/service/config"
	coreRpc "go.amplifyedge.org/sys-share-v2/sys-core/service/go/rpc/v2"
)

const (
	backupFormat = "ver-%s_%s_%s.bak"
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
		// cron backups are unversioned from each other for now.
		_, err := c.backup(fmt.Sprintf("independent-%d", sharedConfig.CurrentTimestamp()))
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

func (c *CoreDB) singleBackup(ctx context.Context, versionPrefix string) (*coreRpc.SingleBackupResult, error) {
	filename, err := c.backup(versionPrefix)
	if err != nil {
		return nil, err
	}
	return &coreRpc.SingleBackupResult{BackupFile: filename}, nil
}

func (a *AllDBService) Backup(ctx context.Context, in *emptypb.Empty) (*coreRpc.BackupAllResult, error) {
	versionPrefix := sharedConfig.NewID()
	var backupFileNames []*coreRpc.SingleBackupResult
	for _, cdb := range a.RegisteredDBs {
		sbr, err := cdb.singleBackup(ctx, versionPrefix)
		if err != nil {
			return nil, err
		}
		backupFileNames = append(backupFileNames, sbr)
	}
	return &coreRpc.BackupAllResult{
		Version:     versionPrefix,
		BackupFiles: backupFileNames,
	}, nil

}

func (c *CoreDB) backup(versionPrefix string) (string, error) {
	c.logger.Debug("creating backup schedule")
	fileWriter, filename, err := c.createBackupFile(versionPrefix)
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

func (c *CoreDB) singleRestore(_ context.Context, in *coreRpc.SingleRestoreRequest) (*coreRpc.SingleRestoreResult, error) {
	badgerDB := c.engine.DB
	f, err := c.openFile(in.BackupFile)
	if err != nil {
		return nil, err
	}
	err = badgerDB.Load(f, 10)
	if err != nil {
		return nil, err
	}
	return &coreRpc.SingleRestoreResult{Result: fmt.Sprintf("successfully restore db: %s", in.BackupFile)}, nil
}

func (a *AllDBService) Restore(ctx context.Context, in *coreRpc.RestoreAllRequest) (*coreRpc.RestoreAllResult, error) {
	if in.RestoreVersion == "" && (in.BackupFiles == nil || len(in.BackupFiles) == 0) {
		return nil, status.Errorf(codes.InvalidArgument, "restore version or specific backup files has to be specified")
	}
	var singleRestoreResults []*coreRpc.SingleRestoreResult
	if in.RestoreVersion != "" {
		for _, cdb := range a.RegisteredDBs {
			backupDir := cdb.config.CronConfig.BackupDir
			backupFilename, err := fileutils.LookupFile(backupDir, in.RestoreVersion)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "restore version %s for database %s not found", in.RestoreVersion, cdb.config.DbConfig.Name)
			}
			res, err := cdb.singleRestore(ctx, &coreRpc.SingleRestoreRequest{BackupFile: backupFilename})
			if err != nil {
				return nil, status.Errorf(codes.Internal, "unable to execute restore version %s for database %s: %v", in.RestoreVersion, cdb.config.DbConfig.Name, err)
			}
			singleRestoreResults = append(singleRestoreResults, res)
		}
		return &coreRpc.RestoreAllResult{RestoreResults: singleRestoreResults}, nil
	}
	if in.BackupFiles != nil && len(in.BackupFiles) != 0 {
		for k, v := range in.BackupFiles {
			cdb := a.FindCoreDB(k)
			if cdb == nil {
				return nil, status.Errorf(codes.InvalidArgument, "unable to find database with name: %s", k)
			}
			res, err := cdb.singleRestore(ctx, &coreRpc.SingleRestoreRequest{BackupFile: v})
			if err != nil {
				return nil, status.Errorf(codes.Internal, "unable to execute restore version %s for database %s: %v", in.RestoreVersion, cdb.config.DbConfig.Name, err)
			}
			singleRestoreResults = append(singleRestoreResults, res)
		}
		return &coreRpc.RestoreAllResult{RestoreResults: singleRestoreResults}, nil
	}
	return nil, status.Errorf(codes.Unknown, "unknown error occured")
}

func (c *CoreDB) SingleListBackup(ctx context.Context, in *emptypb.Empty) ([]*coreRpc.SingleBackupResult, error) {
	var bfiles []*coreRpc.SingleBackupResult
	listFiles, err := c.listBackups()
	if err != nil {
		return nil, err
	}
	for _, f := range listFiles {
		bfiles = append(bfiles, &coreRpc.SingleBackupResult{BackupFile: f})
	}
	return bfiles, nil
}

func (a *AllDBService) ListBackup(ctx context.Context, in *coreRpc.ListBackupRequest) (*coreRpc.ListBackupResult, error) {
	var err error
	backupMaps := map[string][]*coreRpc.SingleBackupResult{}
	var backupAllResults []*coreRpc.BackupAllResult
	if in.BackupVersion != "" {
		backupMaps, err = a.listAndFilterBackups(func(bfile string, version string, bmap map[string][]*coreRpc.SingleBackupResult) error {
			if version == in.BackupVersion {
				bmap[version] = append(bmap[version], &coreRpc.SingleBackupResult{BackupFile: bfile})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		backupMaps, err = a.listAndFilterBackups(func(bfile string, version string, bmap map[string][]*coreRpc.SingleBackupResult) error {
			if version == "" {
				return status.Errorf(codes.Internal, "unable to get version from backups")
			}
			bmap[version] = append(bmap[version], &coreRpc.SingleBackupResult{BackupFile: bfile})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for k, v := range backupMaps {
		backupAllResults = append(backupAllResults, &coreRpc.BackupAllResult{
			Version:     k,
			BackupFiles: v,
		})
	}
	return &coreRpc.ListBackupResult{
		BackupVersions: backupAllResults,
	}, nil
}

func (a *AllDBService) listAndFilterBackups(callbackFunc func(backupFile string, version string, bmap map[string][]*coreRpc.SingleBackupResult) error) (map[string][]*coreRpc.SingleBackupResult, error) {
	backupMaps := map[string][]*coreRpc.SingleBackupResult{}
	for _, cdb := range a.RegisteredDBs {
		blist, err := cdb.listBackups()
		if err != nil {
			return nil, err
		}
		for _, bfile := range blist {
			version := getVersion(bfile)
			if err = callbackFunc(bfile, version, backupMaps); err != nil {
				return nil, err
			}
		}
	}
	return backupMaps, nil
}

func (c *CoreDB) listBackups() ([]string, error) {
	backupDir := c.config.CronConfig.BackupDir
	c.logger.Info("backup dir: " + backupDir)
	fileInfos, err := fileutils.ListFiles(backupDir)
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, f := range fileInfos {
		filenames = append(filenames, f.Name())
	}
	return filenames, nil
}

func (c *CoreDB) createBackupFile(versionPrefix string) (io.WriteCloser, string, error) {
	currentTime := time.Now().Format("200601021859")
	backupFileName := filepath.Join(
		c.config.CronConfig.BackupDir,
		fmt.Sprintf(backupFormat, versionPrefix, c.config.DbConfig.Name, currentTime),
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
	exists := fileutils.FileExists(filepath)
	if !exists {
		return nil, fmt.Errorf("cannot find %s", filepath)
	}
	return os.Open(filepath)
}

func getVersion(filename string) string {
	fslice := strings.Split(filename, "_")
	if len(fslice) == 3 {
		versionSlice := strings.Split(fslice[0], "-")
		if len(versionSlice) == 2 {
			return versionSlice[1]
		}
	}
	return ""
}
