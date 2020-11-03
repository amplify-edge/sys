package db_test

import (
	"fmt"
	commonCfg "github.com/getcouragenow/sys-share/sys-core/service/config/common"
	"testing"

	"github.com/stretchr/testify/assert"

	corecfg "github.com/getcouragenow/sys/sys-core/service/go"
)



func testNewSysCoreConfig(t *testing.T) {
	baseTestDir := "./config"
	// Test nonexistent config
	_, err := corecfg.NewConfig("./nonexistent.yml")
	assert.Error(t, err)
	// Test valid config
	sysCoreCfg, err = corecfg.NewConfig(fmt.Sprintf("%s/%s", baseTestDir, "valid.yml"))
	assert.NoError(t, err)
	expected := &corecfg.SysCoreConfig{
		SysCoreConfig: commonCfg.Config{
			DbConfig: commonCfg.DbConfig{
				Name:             "getcouragenow.db",
				EncryptKey:       "testkey!@",
				RotationDuration: 1,
				DbDir:            "./db",
				DeletePrevious:   true,
			},
			CronConfig: commonCfg.CronConfig{
				BackupSchedule: "@daily",
				RotateSchedule: "@every 3s",
				BackupDir:      "./db/backups",
			},
		},
	}
	assert.Equal(t, expected, sysCoreCfg)
}
