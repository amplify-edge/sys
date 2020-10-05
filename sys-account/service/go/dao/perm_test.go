package dao_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/getcouragenow/sys/sys-account/service/go/dao"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
)

var (
	perms = []*dao.Permission{
		{
			// Admin of an Org
			ID:        role1ID,
			AccountId: accs[0].ID,
			Role:      fmt.Sprintf("%d", 3), // 3 is Admin
			OrgId:     db.UID(),
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: 0,
		},
		{
			// Member of an Org
			ID:        role2ID,
			AccountId: accs[1].ID,
			Role:      fmt.Sprintf("%d", 2), // 2 is member
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

func testPermGet(t *testing.T) {
	t.Log("on querying permission / role")
	perm, err := accdb.GetRole(&dao.QueryParams{Params: map[string]interface{}{
		"id": role1ID,
	}})
	assert.NoError(t, err)
	assert.Equal(t, perm.ID, perms[0].ID)
	t.Logf("Permission queried: %v", perm)
}

func testPermList(t *testing.T) {
	t.Log("on listing / searching permission / role")
	perm, err := accdb.ListRole(&dao.QueryParams{Params: map[string]interface{}{
		"project_id": perms[1].ProjectId,
		"role":       fmt.Sprintf("%d", 2),
	}})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(perm))
	assert.Equal(t, perms[1].Role, perm[0].Role)
	t.Logf("Permission queried: %v", perm)
}

func testPermUpdate(t *testing.T) {
	t.Log("on updating role / permission")
	err := accdb.UpdateRole(&dao.Permission{
		Role:      "3",
		ProjectId: perms[1].ProjectId,
	})
	assert.NoError(t, err)
}
