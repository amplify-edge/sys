package server

import (
	"fmt"

	"github.com/gen0cide/cfx"
)

const (
	ModuleName       = "sys-account"
	errParsingConfig = "error parsing %s config: %v\n"
)

type SysAccountConfig struct {
	UnauthenticatedRoutes []string  `json:"unauthenticatedRoutes,required" yaml:"unauthenticatedRoutes,required" mapstructure:"unauthenticatedRoutes,required"`
	JWTConfig             JWTConfig `json:"jwt,required" yaml:"jwt,required" mapstructure:"jwt,required"`
}

type JWTConfig struct {
	Access  TokenConfig `json:"access,omitempty" yaml:"access,omitempty" mapstructure:"access,omitempty"`
	Refresh TokenConfig `json:"refresh,omitempty" yaml:"refresh,omitempty" mapstructure:"refresh,omitempty"`
}

type TokenConfig struct {
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty" mapstructure:"secret,omitempty"`
	Expiry int    `json:"expiry,omitempty" yaml:"expiry,omitempty" mapstructure:"expiry,omitempty"`
}

func NewConfig(provider cfx.Container) (*SysAccountConfig, error) {
	// create an empty Config object
	cfg := &SysAccountConfig{}

	// use the provider to populate the config
	err := provider.Populate(ModuleName, cfg)
	if err != nil {
		return nil, fmt.Errorf(errParsingConfig, ModuleName, err)
	}

	return cfg, nil
}
