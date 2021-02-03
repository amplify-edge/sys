package dao

import (
	"github.com/amplify-cms/sys-share/sys-core/service/logging"
	"github.com/amplify-cms/sys/sys-core/service/go/pkg/coredb"
)

type FileDB struct {
	db          *coredb.CoreDB
	log         logging.Logger
	fileColumns string
}

func NewFileDB(db *coredb.CoreDB, l logging.Logger) (*FileDB, error) {
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
