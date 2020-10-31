package dao

// const.go contains
// all the table columns / fields basically.
const (
	AccTableName     = "accounts"
	RolesTableName   = "roles"
	OrgTableName     = "orgs"
	ProjectTableName = "projects"
	DefaultLimit     = 50
	DefaultCursor    = `created_at`
)
