package dao

import (
	"strings"
)

// const.go contains
// all the table columns / fields basically.
const (
	AccTableName     = "accounts"
	RolesTableName   = "roles"
	AccColumns       = `id, email, password, role_id, user_defined_fields, survey, created_at, updated_at, last_login, disabled`
	AccColumnsType   = `TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, INTEGER, INTEGER, INTEGER, BOOL`
	AccCursor        = `created_at`
	RolesColumns     = `id, account_id, role, project_id, org_id, created_at, updated_at`
	RolesColumnsType = `TEXT, TEXT, INTEGER, TEXT, TEXT, INTEGER, INTEGER`
	DefaultLimit     = 10
)

// initFields will only be called once during AccountDB initialization (singleton)
func initFields(columns string, values string) map[string]string {
	ret := map[string]string{}
	vals := strings.Split(values, ",")
	for i, col := range strings.Split(columns, ",") {
		if i == 0 {
			ret[col] = vals[i] + " PRIMARY KEY"
		} else {
			ret[col] = vals[i]
		}
	}
	return ret
}
