package dao_test

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/getcouragenow/sys/sys-account/service/go/pkg/dao"
	corecfg "github.com/getcouragenow/sys/sys-core/service/go"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

var (
	testDb     *coresvc.CoreDB
	accdb      *dao.AccountDB
	err        error
	role1ID    = coresvc.NewID()
	role2ID    = coresvc.NewID()
	org1ID     = coresvc.NewID()
	org2ID     = coresvc.NewID()
	proj1ID    = coresvc.NewID()
	proj2ID    = coresvc.NewID()
	proj3ID    = coresvc.NewID()
	account0ID = coresvc.NewID()
	now        = coresvc.CurrentTimestamp()
	accs       = []dao.Account{
		{
			ID:               account0ID,
			Email:            "2pac@example.com",
			Password:         "no_biggie",
			CreatedAt:        now,
			UpdatedAt:        now,
			AvatarResourceId: "https://avatars3.githubusercontent.com/u/59567775?s=200&v=4",
			LastLogin:        0,
			Disabled:         false,
		},
		{
			ID:               coresvc.NewID(),
			Email:            "bigg@example.com",
			Password:         "two_packs",
			CreatedAt:        now,
			UpdatedAt:        now,
			AvatarResourceId: "https://avatars3.githubusercontent.com/u/59567775?s=200&v=3",
			LastLogin:        0,
			Disabled:         false,
		},
		{
			ID:                coresvc.NewID(),
			Email:             "shakur@example.com",
			Password:          "no_biggie",
			AvatarResourceId:  "https://avatars3.githubusercontent.com/u/59567775?s=300&v=4",
			CreatedAt:         now,
			UpdatedAt:         now,
			LastLogin:         0,
			Disabled:          false,
			VerificationToken: "blaharsoaiten",
		},
	}
)

func init() {
	var csc *corecfg.SysCoreConfig
	csc, err = corecfg.NewConfig("./testdata/syscore.yml")
	if err != nil {
		log.Fatalf("error initializing db: %v", err)
	}
	logger := log.New().WithField("test", "sys-account")
	logger.Level = log.DebugLevel
	testDb, err = coresvc.NewCoreDB(logger, &csc.SysCoreConfig, nil)
	if err != nil {
		log.Fatalf("error creating CoreDB: %v", err)
	}
	log.Debug("MakeSchema testing .....")
	accdb, err = dao.NewAccountDB(testDb, logger)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully initialize accountdb :  %v", accdb)
}

func TestAll(t *testing.T) {
	t.Run("Test Account Insert", testAccountInsert)
	t.Run("Test Org Insert", testOrgInsert)
	t.Run("Test Org Get", testOrgGet)
	t.Run("Test Org List", testOrgList)
	t.Run("Test Project Insert", testProjInsert)
	t.Run("Test Project Get", testProjGet)
	t.Run("Test Account Query", testQueryAccounts)
	t.Run("Test Project List", testProjList)
	t.Run("Test Role Insert", testRolesInsert)
	t.Run("Test Role List", testRolesList)
	t.Run("Test Role Get", testRolesGet)
	t.Run("Test Org Update", testUpdateOrg)
	t.Run("Test Project Update", testProjUpdate)
	t.Run("Test Role Update", testRolesUpdate)
	t.Run("Test Account Update", testUpdateAccounts)
	t.Run("Test Org Delete", testDeleteOrg)
	t.Run("Test Account Delete", testDeleteAccounts)
	t.Run("Test Project Delete", testProjDelete)
	t.Run("Test Role Delete", testRoleDelete)
}

func testAccountInsert(t *testing.T) {
	t.Log("on inserting accounts")

	for _, acc := range accs {
		err = accdb.InsertAccount(&acc)
		assert.NoError(t, err)
	}

}

func testQueryAccounts(t *testing.T) {
	t.Logf("on querying accounts")
	queryParams := []*coresvc.QueryParams{
		{
			Params: map[string]interface{}{
				"email": "bigg@example.com",
			},
		},
		{
			Params: map[string]interface{}{
				"email": "2pac@example.com",
			},
		},
	}
	var accs []*dao.Account
	for _, qp := range queryParams {
		acc, err := accdb.GetAccount(qp)
		assert.NoError(t, err)
		t.Logf("Account: %v\n", acc)
		accs = append(accs, acc)
	}
	assert.NotEqual(t, accs[0], accs[1])
	var next int64

	for _, qp := range queryParams {
		accs, next, err = accdb.ListAccount(qp, "email", 1, 0)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, next)
	}
}

func testUpdateAccounts(t *testing.T) {
	accs[0].Email = "makavelli@example.com"
	accs[1].Email = "notorious_big@example.com"
	accs[2].VerificationToken = "MopedRulesTheHighway"

	for _, acc := range accs {
		err = accdb.UpdateAccount(&acc)
		assert.NoError(t, err)
	}

	var getAccounts []*dao.Account

	for _, acc := range accs {
		getAcc, err := accdb.GetAccount(&coresvc.QueryParams{
			Params: map[string]interface{}{"id": acc.ID},
		})
		assert.NoError(t, err)
		getAccounts = append(getAccounts, getAcc)
	}
	assert.Equal(t, accs[0].Email, getAccounts[0].Email)
	assert.Equal(t, accs[1].Email, getAccounts[1].Email)
	assert.Equal(t, accs[2].VerificationToken, getAccounts[2].VerificationToken)
	t.Logf("Updated token: %s", getAccounts[2].VerificationToken)
}

func testDeleteAccounts(t *testing.T) {
	assert.NoError(t, accdb.DeleteAccount(accs[0].ID))
}
