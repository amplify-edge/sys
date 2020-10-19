package service

import (
	"fmt"
	"os"

	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"
	"gopkg.in/yaml.v2"
)

const (
	errParsingConfig = "error parsing %s config: %v\n"
	errDbNameEmpty   = "error: db name empty"
	errDbRotation    = "error: db rotation has to be greater than or equal to 1 (day)"
	errCronSchedule  = "error: db cron schedule is in wrong format / empty"
	defaultDirPerm   = 0755
)

func NewConfig(filepath string) (*SysCoreConfig, error) {
	sysCfg := &SysCoreConfig{}
	f, err := sharedConfig.LoadFile(filepath)
	if err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict(f, &sysCfg); err != nil {
		return nil, fmt.Errorf(errParsingConfig, filepath, err)
	}

	return sysCfg, nil
}

type SysCoreConfig struct {
	SysCoreConfig Config `yaml:"sysCoreConfig" mapstructure:"sysCoreConfig"`
}

func (s *SysCoreConfig) Validate() error {
	return s.SysCoreConfig.validate()
}

type DbConfig struct {
	Name             string `json:"name" yaml:"name" mapstructure:"name"`
	EncryptKey       string `json:"encryptKey" yaml:"encryptKey" mapstructure:"encryptKey"`
	RotationDuration int    `json:"rotationDuration" yaml:"rotationDuration" mapstructure:"rotationDuration"`
	DbDir            string `json:"dbDir" yaml:"dbDir" mapstructure:"dbDir"`
}

func (d DbConfig) validate() error {
	if d.Name == "" {
		return fmt.Errorf(errDbNameEmpty)
	}
	if d.RotationDuration < 1 {
		return fmt.Errorf(errDbRotation)
	}
	if d.EncryptKey == "" {
		encKey, err := sharedConfig.GenRandomByteSlice(32)
		if err != nil {
			return err
		}
		d.EncryptKey = string(encKey)
	}
	exists, err := sharedConfig.PathExists(d.DbDir)
	if err != nil || !exists {
		return os.MkdirAll(d.DbDir, defaultDirPerm)
	}
	return nil
}

type CronConfig struct {
	BackupSchedule string `json:"backupSchedule" yaml:"backupSchedule" mapstructure:"backupSchedule"`
	RotateSchedule string `json:"rotateSchedule" yaml:"rotateSchedule" mapstructure:"rotateSchedule"`
	BackupDir      string `json:"backupDir" yaml:"backupDir" mapstructure:"backupDir"`
}

func (c CronConfig) validate() error {
	if c.BackupSchedule == "" || c.RotateSchedule == "" {
		return fmt.Errorf(errCronSchedule)
	}
	if exists, err := sharedConfig.PathExists(c.BackupDir); err != nil || !exists {
		return os.MkdirAll(c.BackupDir, defaultDirPerm)
	}
	return nil
}

type Config struct {
	DbConfig   DbConfig   `json:"db" yaml:"db" mapstructure:"db"`
	CronConfig CronConfig `json:"cron" yaml:"cron" mapstructure:"cron"`
}

func (c Config) validate() error {
	if err := c.DbConfig.validate(); err != nil {
		return err
	}
	if err := c.CronConfig.validate(); err != nil {
		return err
	}
	return nil

}
