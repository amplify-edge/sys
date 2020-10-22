package dao_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/service"
)

var (
	perms = []*dao.Role{
		{
			// Admin of an Org
			ID:        role1ID,
			AccountId: account0ID,
			Role:      3, // 3 is Admin
			OrgId:     coresvc.NewID(),
			CreatedAt: time.Now().UTC().Unix(),
		},
		{
			// Member of an Org
			ID:        role2ID,
			AccountId: accs[1].ID,
			Role:      2, // 2 is member
			ProjectId: coresvc.NewID(),
			OrgId:     coresvc.NewID(),
			CreatedAt: time.Now().UTC().Unix(),
		},
	}
)

func testRolesInsert(t *testing.T) {
	t.Log("on inserting permissions / roles")
	for _, role := range perms {
		err = accdb.InsertRole(role)
		assert.NoError(t, err)
	}
}

func testRolesGet(t *testing.T) {
	t.Log("on querying permission / role")
	perm, err := accdb.GetRole(&coresvc.QueryParams{Params: map[string]interface{}{
		"id": role1ID,
	}})
	assert.NoError(t, err)
	assert.Equal(t, perms[0], perm)

	perm, err = accdb.GetRole(&coresvc.QueryParams{Params: map[string]interface{}{
		"account_id": account0ID,
	}})
	assert.NoError(t, err)
	assert.Equal(t, perms[0], perm)
}

func testRolesList(t *testing.T) {
	t.Log("on listing / searching permission / role")
	perm, err := accdb.ListRole(&coresvc.QueryParams{Params: map[string]interface{}{
		"project_id": perms[1].ProjectId,
		"role":       2,
	}})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(perm))
	assert.Equal(t, perms[1].Role, perm[0].Role)
	t.Logf("Role queried: %v", perm[0])
}

func testRolesUpdate(t *testing.T) {
	t.Log("on updating role / permission")
	err := accdb.UpdateRole(&dao.Role{
		Role:      3,
		ProjectId: perms[1].ProjectId,
	})
	assert.NoError(t, err)
}
