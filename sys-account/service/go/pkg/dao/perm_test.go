package dao_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
)

var (
	perms = []*dao.Permission{
		{
			// Admin of an Org
			ID:        role1ID,
			AccountId: account0ID,
			Role:      3, // 3 is Admin
			OrgId:     db.UID(),
			CreatedAt: time.Now().UTC().Unix(),
		},
		{
			// Member of an Org
			ID:        role2ID,
			AccountId: accs[1].ID,
			Role:      2, // 2 is member
			ProjectId: db.UID(),
			OrgId:     db.UID(),
			CreatedAt: time.Now().UTC().Unix(),
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
	assert.Equal(t, perms[0], perm)

	perm, err = accdb.GetRole(&dao.QueryParams{Params: map[string]interface{}{
		"account_id": account0ID,
	}})
	assert.NoError(t, err)
	assert.Equal(t, perms[0], perm)
}

func testPermList(t *testing.T) {
	t.Log("on listing / searching permission / role")
	perm, err := accdb.ListRole(&dao.QueryParams{Params: map[string]interface{}{
		"project_id": perms[1].ProjectId,
		"role":       2,
	}})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(perm))
	assert.Equal(t, perms[1].Role, perm[0].Role)
	t.Logf("Permission queried: %v", perm[0])
}

func testPermUpdate(t *testing.T) {
	t.Log("on updating role / permission")
	err := accdb.UpdateRole(&dao.Permission{
		Role:      3,
		ProjectId: perms[1].ProjectId,
	})
	assert.NoError(t, err)
}
