package service

import (
	"fmt"
	"gopkg.in/yaml.v2"

	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"
	commonCfg "github.com/getcouragenow/sys-share/sys-core/service/config/common"
)

const (
	errParsingConfig = "error parsing %s config: %v\n"
)

type SysCoreConfig struct {
	SysCoreConfig commonCfg.Config `yaml:"sysCoreConfig" mapstructure:"sysCoreConfig"`
}

func (s *SysCoreConfig) Validate() error {
	return s.SysCoreConfig.Validate()
}

type DbConfig struct {
	Name             string `json:"name" yaml:"name" mapstructure:"name"`
	EncryptKey       string `json:"encryptKey" yaml:"encryptKey" mapstructure:"encryptKey"`
	RotationDuration int    `json:"rotationDuration" yaml:"rotationDuration" mapstructure:"rotationDuration"`
	DbDir            string `json:"dbDir" yaml:"dbDir" mapstructure:"dbDir"`
	DeletePrevious   bool   `json:"deletePrevious" yaml:"deletePrevious" mapstructure:"deletePrevious"`
}

func NewConfig(filepath string) (*SysCoreConfig, error) {
	sysCfg := &SysCoreConfig{}
	f, err := sharedConfig.LoadFile(filepath)
	if err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict(f, &sysCfg); err != nil {
		return nil, fmt.Errorf(errParsingConfig, filepath, err)
	}
	if err := sysCfg.Validate(); err != nil {
		return nil, err
	}

	return sysCfg, nil
}
