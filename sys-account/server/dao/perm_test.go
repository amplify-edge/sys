package dao_test

import (
	"fmt"
	"github.com/getcouragenow/sys/sys-account/server/dao"
	"github.com/getcouragenow/sys/sys-account/server/rpc/v2"
	"github.com/getcouragenow/sys/sys-core/server/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	perms = []*dao.Permission{
		{
			// Admin of an Org
			ID:        role1ID,
			AccountId: accs[0].ID,
			Role:      fmt.Sprintf("%d", rpc.Roles_ADMIN),
			OrgId:     db.UID(),
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: 0,
		},
		{
			// Member of an Org
			ID:        role2ID,
			AccountId: accs[1].ID,
			Role:      fmt.Sprintf("%d", rpc.Roles_USER),
			ProjectId: db.UID(),
			OrgId:     db.UID(),
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: 0,
		},
	}
)

func testPermInsert(t *testing.T) {
	t.Log("on inserting permissions / roles")
	for _, role := range perms {
		err = accdb.InsertRole(role)
		assert.NoError(t, err)
	}
}
