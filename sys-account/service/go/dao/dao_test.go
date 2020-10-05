package dao_test

import (
	"github.com/getcouragenow/sys/sys-account/service/go/dao"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/genjidb/genji"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/db"
)

var (
	testDb  *genji.DB
	accdb   *dao.AccountDB
	err     error
	role1ID = db.UID()
	role2ID = db.UID()

	accs = []dao.Account{
		{
			ID:       db.UID(),
			Email:    "2pac@example.com",
			Password: "no_biggie",
			RoleId:   role1ID,
			UserDefinedFields: map[string]interface{}{
				"City": "Compton",
			},
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: time.Now().UTC().Unix(),
			LastLogin: 0,
			Disabled:  false,
		},
		{
			ID:       db.UID(),
			Email:    "bigg@example.com",
			Password: "two_packs",
			RoleId:   role2ID,
			UserDefinedFields: map[string]interface{}{
				"City": "NY",
			},
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: time.Now().UTC().Unix(),
			LastLogin: 0,
			Disabled:  false,
		},
		{
			ID:       db.UID(),
			Email:    "2pac@example.com",
			Password: "no_biggie",
			RoleId:   role1ID,
			UserDefinedFields: map[string]interface{}{
				"City": "Compton",
			},
			CreatedAt: time.Now().UTC().Unix(),
			UpdatedAt: time.Now().UTC().Unix(),
			LastLogin: 0,
			Disabled:  false,
		},
	}
)

func init() {
	testDb = db.SharedDatabase()
	log.Println("MakeSchema testing .....")
	accdb, err = dao.NewAccountDB(testDb)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully initialize accountdb :  %v", accdb)
}

func TestAll(t *testing.T) {
	t.Run("Test Account Insert", testAccountInsert)
	t.Run("Test Role Insert", testPermInsert)
	t.Run("Test Role Get", testPermGet)
	t.Run("Test Role List", testPermList)
	t.Run("Test Role Update", testPermUpdate)
	t.Run("Test Account Query", testQueryAccounts)
	t.Run("Test Account Update", testUpdateAccounts)
}

func testAccountInsert(t *testing.T) {
	t.Log("on inserting accounts")

	for _, acc := range accs {
		err = accdb.InsertAccount(&acc)
		assert.NoError(t, err)
	}
	t.Log("successfully inserted accounts")
}

func testQueryAccounts(t *testing.T) {
	t.Logf("on querying accounts")
	queryParams := []*dao.QueryParams{
		{
			Params: map[string]interface{}{
				"name": "Biggie",
			},
		},
		{
			Params: map[string]interface{}{
				"name": "Tupac Shakur",
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

	for _, qp := range queryParams {
		accs, err = accdb.ListAccount(qp)
		assert.NoError(t, err)
	}
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
