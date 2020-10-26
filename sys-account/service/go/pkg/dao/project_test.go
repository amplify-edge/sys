package dao_test

import (
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
)

var (
	projects = []*dao.Project{
		{
			Id:        proj1ID,
			Name:      "Project 1 Org 1",
			LogoUrl:   "https://avatars3.githubusercontent.com/u/59567775?s=200&v=4",
			CreatedAt: 1603520049,
			AccountId: account0ID,
			OrgId:     org1ID,
		},
		{
			Id:        proj2ID,
			Name:      "Project 2 Org 1",
			LogoUrl:   "https://avatars3.githubusercontent.com/u/59567775?s=200&v=4",
			CreatedAt: 1603520089,
			AccountId: account0ID,
			OrgId:     org1ID,
		},
		{
			Id:        proj3ID,
			Name:      "Project 1 Org 2",
			LogoUrl:   "https://avatars3.githubusercontent.com/u/59567775?s=200&v=4",
			CreatedAt: 1603520780,
			AccountId: account0ID,
			OrgId:     org2ID,
		},
	}
)

func testProjInsert(t *testing.T) {
	t.Log("on inserting project")
	for _, proj := range projects {
		err = accdb.InsertProject(proj)
		assert.NoError(t, err)
	}
}

func testProjGet(t *testing.T) {
	t.Log("on get project")
	proj, err := accdb.GetProject(&coresvc.QueryParams{Params: map[string]interface{}{
		"id":         proj1ID,
		"created_at": 1603520049,
	}})
	assert.NoError(t, err)
	assert.Equal(t, projects[0], proj)
}

func testProjList(t *testing.T) {
	t.Log("on list project")
	qps := []*coresvc.QueryParams{
		{
			Params: map[string]interface{}{
				"org_id": org1ID,
			},
		},
		{
			Params: map[string]interface{}{
				"org_id": org2ID,
			},
		},
	}
	var allProjects [][]*dao.Project
	for _, qp := range qps {
		projs, next, err := accdb.ListProject(qp, "name", 2, 0)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, next)
		assert.NotEqual(t, nil, projs)
		allProjects = append(allProjects, projs)
	}
	// assert.Equal(t, allProjects[0], projects[:1])
	assert.Equal(t, allProjects[1][0], projects[2])
	t.Log(allProjects)
}

func testProjUpdate(t *testing.T) {
	projects[0].Name = "Project Uno Org Uno"
	projects[2].Name = "Project Uno Org Dos"

	for _, proj := range projects {
		err = accdb.UpdateProject(proj)
		assert.NoError(t, err)
	}
}

func testProjDelete(t *testing.T) {
	assert.NoError(t, accdb.DeleteProject(projects[0].Id))
}
