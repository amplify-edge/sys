package dao

import (
	"fmt"
)

func delAllStmt(tblName string) string {
	return fmt.Sprintf("DELETE FROM %s", tblName)
}

func (a *AccountDB) ResetAll() error {
	deleteAllProjectsStmt := delAllStmt(ProjectTableName)
	deleteAllOrgsStmt := delAllStmt(OrgTableName)
	deleteAllRoles := fmt.Sprintf("DELETE FROM %s", RolesTableName)
	deleteAllAccounts := fmt.Sprintf("DELETE FROM %s", AccTableName)

	return a.db.BulkExec(map[string][]interface{}{
		deleteAllProjectsStmt: nil,
		deleteAllOrgsStmt:     nil,
		deleteAllRoles:        []interface{}{},
		deleteAllAccounts:     []interface{}{},
	})
}
