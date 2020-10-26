package dao_test

import (
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
)

var (
	orgs = []*dao.Org{
		{
			Id:        org1ID,
			Name:      "Org 1",
			LogoUrl:   "https://avatars3.githubusercontent.com/u/59567775?s=200&v=4",
			Contact:   "contact@example.com",
			CreatedAt: 1603520049,
			AccountId: account0ID,
		},
		{
			Id:        org2ID,
			Name:      "Org 2",
			LogoUrl:   "https://avatars3.githubusercontent.com/u/59567775?s=200&v=4",
			Contact:   "contact2@example.com",
			CreatedAt: 1603520780,
			AccountId: account0ID,
		},
	}
)

func testOrgInsert(t *testing.T) {
	t.Log("on inserting org")
	for _, org := range orgs {
		err = accdb.InsertOrg(org)
		assert.NoError(t, err)
	}
}

func testOrgGet(t *testing.T) {
	t.Log("on get org")
	org, err := accdb.GetOrg(&coresvc.QueryParams{Params: map[string]interface{}{
		"id":         org1ID,
		"created_at": 1603520049,
	}})
	assert.NoError(t, err)
	assert.Equal(t, orgs[0], org)
}

func testOrgList(t *testing.T) {
	t.Log("on list org")
	qps := []*coresvc.QueryParams{
		{
			Params: map[string]interface{}{
				"account_id": account0ID,
			},
		},
		{
			Params: map[string]interface{}{
				"name": "Org 1",
			},
		},
	}
	var allOrgs [][]*dao.Org
	for _, qp := range qps {
		orgs, next, err := accdb.ListOrg(qp, "name", 2, 0)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, next)
		assert.NotEqual(t, nil, orgs)
		allOrgs = append(allOrgs, orgs)
	}
	// assert.Equal(t, allOrgs[0], orgs)
	// assert.Equal(t, orgs[0], allOrgs[1][0])
	t.Log(allOrgs)
}

func testUpdateOrg(t *testing.T) {
	orgs[0].Name = "ORG 1 BRUH"
	orgs[1].Name = "ORG 2 BRUH"

	for _, org := range orgs {
		err = accdb.UpdateOrg(org)
		assert.NoError(t, err)
	}

	var getOrgs []*dao.Org

	for _, org := range orgs {
		getOrg, err := accdb.GetOrg(&coresvc.QueryParams{Params: map[string]interface{}{"id": org.Id}})
		assert.NoError(t, err)
		getOrgs = append(getOrgs, getOrg)
	}
	assert.Equal(t, orgs[0].Name, getOrgs[0].Name)
}

func testDeleteOrg(t *testing.T) {
	assert.NoError(t, accdb.DeleteOrg(orgs[0].Id))
}
