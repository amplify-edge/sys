package dao_test

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

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
	account0ID = coresvc.NewID()
	accs       = []dao.Account{
		{
			ID:       account0ID,
			Email:    "2pac@example.com",
			Password: "no_biggie",
			RoleId:   role1ID,
			Survey:   map[string]interface{}{},
			UserDefinedFields: map[string]interface{}{
				"City": "Compton",
			},
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: time.Now().UTC().Unix(),
			LastLogin: 0,
			Disabled:  false,
		},
		{
			ID:       coresvc.NewID(),
			Email:    "bigg@example.com",
			Password: "two_packs",
			RoleId:   role2ID,
			Survey:   map[string]interface{}{},
			UserDefinedFields: map[string]interface{}{
				"City": "NY",
			},
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: time.Now().UTC().Unix(),
			LastLogin: 0,
			Disabled:  false,
		},
		{
			ID:       coresvc.NewID(),
			Email:    "shakur@example.com",
			Password: "no_biggie",
			RoleId:   role1ID,
			Survey:   map[string]interface{}{},
			UserDefinedFields: map[string]interface{}{
				"City": "Compton LA",
			},
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: time.Now().UTC().Unix(),
			LastLogin: 0,
			Disabled:  false,
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
	testDb, err = coresvc.NewCoreDB(logger, csc)
	if err != nil {
		log.Fatalf("error creating CoreDB: %v", err)
	}
	log.Debug("MakeSchema testing .....")
	accdb, err = dao.NewAccountDB(testDb)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully initialize accountdb :  %v", accdb)
}

func TestAll(t *testing.T) {
	t.Run("Test Account Insert", testAccountInsert)
	t.Run("Test Role Insert", testRolesInsert)
	t.Run("Test Account Query", testQueryAccounts)
	t.Run("Test Role List", testRolesList)
	t.Run("Test Role Get", testRolesGet)
	t.Run("Test Role Update", testRolesUpdate)
	t.Run("Test Account Update", testUpdateAccounts)
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
	}
	assert.NotEqual(t, 0, next)
}

func testUpdateAccounts(t *testing.T) {
	accs[0].Email = "makavelli@example.com"
	accs[1].Email = "notorious_big@example.com"

	for _, acc := range accs {
		err = accdb.UpdateAccount(&acc)
		assert.NoError(t, err)
	}
}

func testDeleteAccounts(t *testing.T) {
	assert.NoError(t, accdb.DeleteAccount(accs[0].ID))
}
