package filesvc

import (
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc/dao"
	"github.com/sirupsen/logrus"
)

type SysFileSvc struct {
	store *dao.FileDB
	log   *logrus.Entry
}

func NewSysFileSvc(db *coredb.CoreDB, log *logrus.Entry) (*SysFileSvc, error) {
	store, err := dao.NewFileDB(db, log)
	if err != nil {
		return nil, err
	}
	return &SysFileSvc{
		store: store,
		log:   log,
	}, nil
}
