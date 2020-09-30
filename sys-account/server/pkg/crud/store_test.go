package crud_test

import (
	"github.com/getcouragenow/sys/sys-account/server/pkg/crud"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTableStmt(t *testing.T) {
	tbl := crud.NewTable("accounts", map[string]string{
		"id": "TEXT PRIMARY KEY",
		"blah_id": "TEXT UNIQUE",
		"email": "TEXT",
		"jsonField": "BLOB",
	})
	ret := tbl.CreateTable()
	t.Logf("create table statement: %s", ret)
	assert.NotEqual(t, "", ret)
	assert.NotContains(t, "jsonField", ret)
}
