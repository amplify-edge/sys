package filesvc

import (
	"github.com/getcouragenow/sys-share/sys-core/service/logging"
	"google.golang.org/grpc"

	sharedPkg "github.com/getcouragenow/sys-share/sys-core/service/go/rpc/v2"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc/repo"
)

type SysFileService struct {
	repo *repo.SysFileRepo
}

func NewSysFileService(cfg *FileServiceConfig, l logging.Logger) (*SysFileService, error) {
	db, err := coredb.NewCoreDB(l, &cfg.DBConfig, nil)
	if err != nil {
		return nil, err
	}
	fileRepo, err := repo.NewSysFileRepo(db, l)
	if err != nil {
		return nil, err
	}
	return &SysFileService{repo: fileRepo}, nil
}

func (s *SysFileService) RegisterService(srv *grpc.Server) {
	sharedPkg.RegisterFileServiceService(srv, sharedPkg.NewFileServiceService(s.repo))
}
