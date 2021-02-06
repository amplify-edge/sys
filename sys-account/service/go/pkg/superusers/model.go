package superusers

import (
	"go.amplifyedge.org/sys-share-v2/sys-core/service/config"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/logging"
	"go.amplifyedge.org/sys-v2/sys-account/service/go/pkg/pass"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultSuperAdmin         = "superadmin"
	defaultSuperadminPassword = "superadmin"
	defaultAvatar             = `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`
	defaultSuperAdminFilepath = "./config/supers.yml"
	defaultFilePerm           = 0660
)

type SuperUser struct {
	// SuperUser has names as opposed to email
	Name string
	// HashedPassword contains hashed password
	HashedPassword string
	// Avatar is a base64
	Avatar string
}

type SuperUserConfig struct {
	SuperUsers []*SuperUser `json:"super_users" yaml:"super_users" mapstructure:"super_users"`
}

// SuperUserIO queries superusers from a config under yaml / json format
// note that passwords are and should be encrypted
type SuperUserIO struct {
	fpath    string
	logger   logging.Logger
	mu       sync.RWMutex
}

func (s *SuperUserIO) GetFilepath() string {
	return s.fpath
}

// NewSuperUserDAO creates SuperUserIO
// it checks whether or not the file exists, if it does it opens it readonly
// if it doesn't it creates a file containing default superuser and its password in a hashed format.
func NewSuperUserDAO(superAdminFilePath string, logger logging.Logger) *SuperUserIO {
	if superAdminFilePath == "" {
		superAdminFilePath = defaultSuperAdminFilepath
	}
	superDir := filepath.Dir(superAdminFilePath)
	_ = os.MkdirAll(superDir, os.ModeDir)
	mu := sync.RWMutex{}
	f, err := os.OpenFile(superAdminFilePath, os.O_RDONLY, defaultFilePerm)
	if err != nil {
		logger.Error("error opening superadmin config, creating one")
		f, err = os.OpenFile(superAdminFilePath, os.O_CREATE|os.O_WRONLY, defaultFilePerm)
		if err != nil {
			logger.Fatalf("error creating superadmin config: %v", err)
		}
	}
	defer f.Close()
	finfo, _ := f.Stat()
	if finfo.Size() == 0 {
		superHash, err := pass.GenHash(defaultSuperadminPassword)
		if err != nil {
			logger.Fatalf("error generating hash for default superadmin: %v", err)
		}
		hcodedSuper := SuperUserConfig{SuperUsers: []*SuperUser{
			{
				Name:           DefaultSuperAdmin,
				HashedPassword: superHash,
				Avatar:         defaultAvatar,
			},
		}}
		b, err := config.MarshalYAML(&hcodedSuper)
		if err != nil {
			logger.Fatal("unable to marshal default superuser to yml")
		}
		mu.Lock()
		f.Write(b)
		mu.Unlock()
	}
	return &SuperUserIO{
		fpath:    superAdminFilePath,
		logger:   logger,
		mu:       mu,
	}
}
