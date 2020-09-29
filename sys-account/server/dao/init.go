package dao

import "github.com/getcouragenow/sys/sys-core/server/pkg/db"

var (
	tablePrefix = "sys_core"
	modName     = "accounts"
)

func tableName(name string) string {
	return tablePrefix + "_" + modName + "_" + name
}

func init() {
	tables := []db.DbModel{
		User{},
		Project{},
		Org{},
		Roles{},
		Permission{},
	}
	db.RegisterModels(modName, tables)
}
