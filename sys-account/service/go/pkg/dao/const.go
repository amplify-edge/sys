package dao

import (
	"strings"
)

// const.go contains
// all the table columns / fields basically.
const (
	AccTableName       = "accounts"
	RolesTableName     = "roles"
	OrgTableName       = "orgs"
	ProjectTableName   = "projects"
	AccColumns         = `id, email, password, role_id, user_defined_fields, survey, created_at, updated_at, last_login, disabled, verified`
	AccColumnsType     = `TEXT, TEXT, TEXT, TEXT, TEXT, TEXT, INTEGER, INTEGER, INTEGER, BOOL, BOOL`
	AccCursor          = `created_at`
	RolesColumns       = `id, account_id, role, project_id, org_id, created_at, updated_at`
	RolesColumnsType   = `TEXT, TEXT, INTEGER, TEXT, TEXT, INTEGER, INTEGER`
	OrgColumns         = `id, name, logo_url, contact, created_at, account_id`
	OrgColumnsType     = `TEXT, TEXT, TEXT, TEXT, INTEGER, TEXT`
	ProjectColumns     = `id, name, logo_url, created_at, account_id, org_id`
	ProjectColumnsType = `TEXT, TEXT, TEXT, INTEGER, TEXT, TEXT`
	DefaultLimit       = 50
	DefaultCursor      = `created_at`
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
