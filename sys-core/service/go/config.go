package service

import (
	"fmt"
	"github.com/amplify-cms/sys-share/sys-core/service/fileutils"
	"gopkg.in/yaml.v2"

	commonCfg "github.com/amplify-cms/sys-share/sys-core/service/config/common"
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
	SenderName     string     `json:"senderName,omitempty" yaml:"senderName"`
	SenderMail     string     `json:"senderMail,omitempty" yaml:"senderMail"`
	ProductName    string     `json:"productName,omitempty" yaml:"productName"`
	LogoUrl        string     `json:"logoUrl,omitempty" yaml:"logoUrl"`
	Copyright      string     `json:"copyright,omitempty" yaml:"copyright"`
	TroubleContact string     `json:"troubleContact,omitempty" yaml:"troubleContact"`
	Sendgrid       Sendgrid   `json:"sendgrid,omitempty" yaml:"sendgrid"`
	Smtp           SmtpConfig `json:"smtp,omitempty" yaml:"smtp"`
}

type Sendgrid struct {
	ApiKey string `json:"apiKey,omitempty" yaml:"apiKey,omitempty"`
}

func (s *Sendgrid) validate() error {
	// if s.ApiKey == "" {
	// 	return fmt.Errorf(errParsingConfig, "no sendgrid api key provided")
	// }
	return nil
}

type SmtpConfig struct {
	Host     string `json:"host,omitempty" yaml:"host,omitempty"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

func (s *SmtpConfig) validate() error {
	if s.Host == "" {
		return fmt.Errorf(errParsingConfig, "smtp host is empty")
	}
	if s.Port == 0 {
		s.Port = 587
	}
	if s.Email == "" {
		return fmt.Errorf(errParsingConfig, "smtp email is empty")
	}
	if s.Password == "" {
		return fmt.Errorf(errParsingConfig, "smtp password is empty")
	}
	return nil
}

func (m *MailConfig) Validate() error {
	if &m.Smtp != nil {
		return m.Smtp.validate()
	}
	if &m.Sendgrid != nil {
		return m.Sendgrid.validate()
	}
	return nil
}

func NewConfig(filepath string) (*SysCoreConfig, error) {
	sysCfg := &SysCoreConfig{}
	f, err := fileutils.LoadFile(filepath)
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
