package dao

import (
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/service"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	tablePrefix = "sys_accounts"
	modName     = "accounts"
)

type AccountDB struct {
	db  *coresvc.CoreDB
	log *logrus.Logger
}

func NewAccountDB(db *coresvc.CoreDB) (*AccountDB, error) {
	err := db.RegisterModels(map[string]coresvc.DbModel{
		AccTableName: Account{},
		RolesTableName: Role{},
	})
	if err != nil {
		return nil, err
	}
	if err := db.MakeSchema(); err != nil {
		return nil, err
	}
	log := logrus.New()
	return &AccountDB{db, log}, nil
}

func (a *AccountDB) BuildSearchQuery(qs string) string {
	var sb strings.Builder
	sb.WriteString("%")
	sb.WriteString(qs)
	sb.WriteString("%")
	return sb.String()
}
