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
	MailConfig    MailConfig       `yaml:"mailConfig" mapstructure:"mailConfig"`
}

func (s *SysCoreConfig) Validate() error {
	if err := s.MailConfig.Validate(); err != nil {
		return err
	}
	return s.SysCoreConfig.Validate()
}

type DbConfig struct {
	Name             string `json:"name" yaml:"name" mapstructure:"name"`
	EncryptKey       string `json:"encryptKey" yaml:"encryptKey" mapstructure:"encryptKey"`
	RotationDuration int    `json:"rotationDuration" yaml:"rotationDuration" mapstructure:"rotationDuration"`
	DbDir            string `json:"dbDir" yaml:"dbDir" mapstructure:"dbDir"`
	DeletePrevious   bool   `json:"deletePrevious" yaml:"deletePrevious" mapstructure:"deletePrevious"`
}

type MailConfig struct {
	SendgridApiKey string `json:"sendgridApiKey,omitempty" yaml:"sendgridApiKey"`
	SenderName     string `json:"senderName,omitempty" yaml:"senderName"`
	SenderMail     string `json:"senderMail,omitempty" yaml:"senderMail"`
	ProductName    string `json:"productName,omitempty" yaml:"productName"`
	LogoUrl        string `json:"logoUrl,omitempty" yaml:"logoUrl"`
	Copyright      string `json:"copyright,omitempty" yaml:"copyright"`
	TroubleContact string `json:"troubleContact,omitempty" yaml:"troubleContact"`
}

func (m *MailConfig) Validate() error {
	if m.SendgridApiKey == "" {
		return fmt.Errorf(errParsingConfig, "no sendgrid api key provided")
	}
	return nil
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
