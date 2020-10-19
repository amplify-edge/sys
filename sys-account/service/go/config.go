package service

import (
	"fmt"
	sharedConfig "github.com/getcouragenow/sys-share/sys-core/service/config"

	"gopkg.in/yaml.v2"
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
	if err := c.JWTConfig.validate(); err != nil {
		return err
	}
	return nil
}

type JWTConfig struct {
	Access  TokenConfig `json:"access" yaml:"access" mapstructure:"access"`
	Refresh TokenConfig `json:"refresh" yaml:"refresh" mapstructure:"refresh"`
}

func (j JWTConfig) validate() error {
	if err := j.Access.validate(); err != nil {
		return err
	}
	if err := j.Refresh.validate(); err != nil {
		return err
	}
	return nil
}

type TokenConfig struct {
	Secret string `json:"secret" yaml:"secret" mapstructure:"secret"`
	Expiry int    `json:"expiry" yaml:"expiry" mapstructure:"expiry"`
}

func (t TokenConfig) validate() error {
	if t.Secret == "" {
		secret, err := sharedConfig.GenRandomByteSlice(32)
		if err != nil {
			return err
		}
		t.Secret = string(secret)
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

