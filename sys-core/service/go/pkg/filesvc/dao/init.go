package dao

import (
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"github.com/sirupsen/logrus"
)

type FileDB struct {
	db          *coredb.CoreDB
	log         *logrus.Entry
	fileColumns string
}

func NewFileDB(db *coredb.CoreDB, l *logrus.Entry) (*FileDB, error) {
	fileColumns := coredb.GetStructColumns(File{})
	err := db.RegisterModels(map[string]coredb.DbModel{
		FilesTableName: File{},
	})
	if err != nil {
		return nil, err
	}
	if err = db.MakeSchema(); err != nil {
		return nil, err
	}
	return &FileDB{
		db:          db,
		log:         l,
		fileColumns: fileColumns,
	}, nil
}
