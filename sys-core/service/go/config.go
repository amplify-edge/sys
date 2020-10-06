package service

import (
	"fmt"

	"github.com/gen0cide/cfx"
)

const (
	ModuleName       = "sys-core"
	errParsingConfig = "error parsing %s config: %v\n"
)

// NewFooConfig is a constructor that uses the cfx package to inject the configuration
// from CFX parsed YAML.
func NewConfig(provider cfx.Container) (*SysCoreConfig, error) {
	// create an empty Config object
	cfg := &SysCoreConfig{}

	// use the provider to populate the config
	err := provider.Populate(ModuleName, cfg)
	if err != nil {
		return nil, fmt.Errorf(errParsingConfig, ModuleName, err)
	}

	return cfg, nil
}

type DbConfig struct {
	Name             string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`
	EncryptKey       string `json:"encryptKey,omitempty" yaml:"encryptKey,omitempty" mapstructure:"encryptKey,omitempty"`
	RotationDuration int    `json:"rotationDuration,omitempty" yaml:"rotationDuration,omitempty" mapstructure:"rotationDuration,omitempty"`
	DbDir            string `json:"dbDir,omitempty" yaml:"dbDir,omitempty" mapstructure:"dbDir,omitempty"`
}

type CronConfig struct {
	BackupSchedule string `json:"backupSchedule,omitempty" yaml:"backupSchedule,omitempty" mapstructure:"backupSchedule,omitempty"`
	RotateSchedule string `json:"rotateSchedule,omitempty" yaml:"rotateSchedule,omitempty" mapstructure:"rotateSchedule,omitempty"`
	BackupDir      string `json:"backupDir,omitempty" yaml:"backupDir,omitempty" mapstructure:"backupDir,omitempty"`
}

type SysCoreConfig struct {
	DbConfig   DbConfig   `json:"db,required" yaml:"db,required" mapstructure:"db,required"`
	CronConfig CronConfig `json:"cron,required" yaml:"cron,required" mapstructure:"cron,required"`
}
