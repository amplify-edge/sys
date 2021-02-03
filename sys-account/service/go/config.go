package service

import (
	"fmt"
	"gopkg.in/yaml.v2"

	sharedConfig "github.com/amplify-cms/sys-share/sys-core/service/config"
	commonCfg "github.com/amplify-cms/sys-share/sys-core/service/config/common"
	coresvc "github.com/amplify-cms/sys/sys-core/service/go"
)

const (
	errParsingConfig           = "error parsing %s config: %v\n"
	errNoUnauthenticatedRoutes = "error: no unauthenticated routes defined"
)

type SysAccountConfig struct {
	SysAccountConfig Config `yaml:"sysAccountConfig" mapstructure:"sysAccountConfig"`
}

func (s *SysAccountConfig) Validate() error {
	return s.SysAccountConfig.validate()
}

type Config struct {
	SuperUserFilePath     string             `json:"superUserFilePath" yaml:"superUserFilePath" mapstructure:"superUserFilePath"`
	UnauthenticatedRoutes []string           `json:"unauthenticatedRoutes" yaml:"unauthenticatedRoutes" mapstructure:"unauthenticatedRoutes"`
	JWTConfig             JWTConfig          `json:"jwt" yaml:"jwt" mapstructure:"jwt"`
	SysCoreConfig         commonCfg.Config   `yaml:"sysCoreConfig" mapstructure:"sysCoreConfig"`
	SysFileConfig         commonCfg.Config   `yaml:"sysFileConfig" mapstructure:"sysFileConfig"`
	MailConfig            coresvc.MailConfig `yaml:"mailConfig" mapstructure:"mailConfig"`
}

func (c Config) validate() error {
	if len(c.UnauthenticatedRoutes) == 0 {
		return fmt.Errorf(errNoUnauthenticatedRoutes)
	}
	if err := c.JWTConfig.Validate(); err != nil {
		return err
	}
	if err := c.MailConfig.Validate(); err != nil {
		return err
	}
	if err := c.SysCoreConfig.Validate(); err != nil {
		return err
	}
	if err := c.SysFileConfig.Validate(); err != nil {
		return err
	}
	return nil
}

type JWTConfig struct {
	Access  commonCfg.TokenConfig `json:"access" yaml:"access" mapstructure:"access"`
	Refresh commonCfg.TokenConfig `json:"refresh" yaml:"refresh" mapstructure:"refresh"`
}

func (j JWTConfig) Validate() error {
	if err := j.Access.Validate(); err != nil {
		return err
	}
	if err := j.Refresh.Validate(); err != nil {
		return err
	}
	return nil
}

func NewConfig(filepath string) (*SysAccountConfig, error) {
	cfg := &SysAccountConfig{}
	f, err := sharedConfig.LoadFile(filepath)
	if err != nil {
		return nil, err
	}
	if err = yaml.UnmarshalStrict(f, &cfg); err != nil {
		return nil, fmt.Errorf(errParsingConfig, filepath, err)
	}
	if err = cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
