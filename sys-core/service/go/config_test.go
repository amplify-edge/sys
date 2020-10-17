package service_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	corecfg "github.com/getcouragenow/sys/sys-core/service/go"
)

func TestNewSysCoreConfig(t *testing.T) {
	baseTestDir := "./test/config"
	// Test nonexistent config
	_, err := corecfg.NewConfig("./nonexistent.yml")
	assert.Error(t, err)
	// Test valid config
	sysCoreCfg, err := corecfg.NewConfig(fmt.Sprintf("%s/%s", baseTestDir, "valid.yml"))
	assert.NoError(t, err)
	expected := &corecfg.SysCoreConfig{
		SysCoreConfig: corecfg.Config{
			DbConfig: corecfg.DbConfig{
				Name:             "getcouragenow.db",
				EncryptKey:       "testkey!@",
				RotationDuration: 1,
				DbDir:            "./db",
			},
			CronConfig: corecfg.CronConfig{
				BackupSchedule: "@daily",
				RotateSchedule: "@every 24h",
				BackupDir:      "./db/backups",
			},
		},
	}
	assert.Equal(t, expected, sysCoreCfg)
}
