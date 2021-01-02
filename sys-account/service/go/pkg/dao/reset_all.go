package dao

import (
	"fmt"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"strings"
)

func delAllStmt(tblName string) string {
	return fmt.Sprintf("DELETE FROM %s", tblName)
}

func (a *AccountDB) ResetAll(initialSuperUsersMail []string) error {
	deleteAllProjectsStmt := delAllStmt(ProjectTableName)
	deleteAllOrgsStmt := delAllStmt(OrgTableName)
	var accountIds []string
	var roleIds []string
	for _, mail := range initialSuperUsersMail {
		acc, err := a.GetAccount(&coresvc.QueryParams{Params: map[string]interface{}{
			"email": mail,
		}})
		if err != nil {
			return err
		}
		accountIds = append(accountIds, acc.ID)
		roles, err := a.FetchRoles(acc.ID)
		if err != nil {
			return err
		}
		for _, r := range roles {
			roleIds = append(roleIds, r.ID)
		}
	}
	joinedRoleIds := strings.Join(roleIds, ",")
	joinedAccountIds := strings.Join(accountIds, ",")
	deleteAllRoles := fmt.Sprintf("DELETE FROM %s WHERE id NOT IN [?]", RolesTableName)
	deleteAllAccounts := fmt.Sprintf("DELETE FROM %s WHERE id NOT IN [?]", AccTableName)

	return a.db.BulkExec(map[string][]interface{}{
		deleteAllProjectsStmt: nil,
		deleteAllOrgsStmt:     nil,
		deleteAllRoles:        []interface{}{joinedRoleIds},
		deleteAllAccounts:     []interface{}{joinedAccountIds},
	})
}
