package dao

import (
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"github.com/sirupsen/logrus"
	"strings"
)

type AccountDB struct {
	db                  *coresvc.CoreDB
	log                 *logrus.Entry
	accountColumns      string
	roleColumns         string
	orgColumns          string
	projectColumns      string
	loginAttemptColumns string
}

func NewAccountDB(db *coresvc.CoreDB, l *logrus.Entry) (*AccountDB, error) {
	accColumns := coresvc.GetStructColumns(Account{})
	roleColumns := coresvc.GetStructColumns(Role{})
	orgColumns := coresvc.GetStructColumns(Org{})
	projectColumns := coresvc.GetStructColumns(Project{})
	loginAttemptColumns := coresvc.GetStructColumns(LoginAttempt{})

	err := db.RegisterModels(map[string]coresvc.DbModel{
		AccTableName:        Account{},
		RolesTableName:      Role{},
		OrgTableName:        Org{},
		ProjectTableName:    Project{},
		loginAttemptColumns: LoginAttempt{},
	})
	if err != nil {
		return nil, err
	}
	if err := db.MakeSchema(); err != nil {
		return nil, err
	}
	return &AccountDB{db, l, accColumns, roleColumns, orgColumns, projectColumns, loginAttemptColumns}, nil
}

func (a *AccountDB) BuildSearchQuery(qs string) string {
	var sb strings.Builder
	sb.WriteString("%")
	sb.WriteString(qs)
	sb.WriteString("%")
	return sb.String()
}
