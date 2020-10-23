package service

import (
	"fmt"
	"gopkg.in/yaml.v2"

	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"
	commonCfg "github.com/getcouragenow/sys-share/sys-core/service/config/common"
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
	UnauthenticatedRoutes []string  `json:"unauthenticatedRoutes" yaml:"unauthenticatedRoutes" mapstructure:"unauthenticatedRoutes"`
	JWTConfig             JWTConfig `json:"jwt" yaml:"jwt" mapstructure:"jwt"`
}

func (c Config) validate() error {
	if len(c.UnauthenticatedRoutes) == 0 {
		return fmt.Errorf(errNoUnauthenticatedRoutes)
	}
	if err := c.JWTConfig.Validate(); err != nil {
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
	if err := yaml.UnmarshalStrict(f, &cfg); err != nil {
		return nil, fmt.Errorf(errParsingConfig, filepath, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
