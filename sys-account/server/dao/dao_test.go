package dao_test

import (
	"github.com/getcouragenow/sys/sys-account/server/dao"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/genjidb/genji"
	"github.com/getcouragenow/sys/sys-core/server/pkg/db"
)

var (
	testDb *genji.DB
	accdb  *dao.AccountDB
	err    error

	accs = []dao.Account{
		{
			ID:       db.UID(),
			Name:     "Tupac Shakur",
			Email:    "2pac@example.com",
			Password: "no_biggie",
			RoleId:   db.UID(),
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
			Name:     "Biggie",
			Email:    "bigg@example.com",
			Password: "two_packs",
			RoleId:   db.UID(),
			UserDefinedFields: map[string]interface{}{
				"City": "NY",
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
	t.Run("testAccountInsert", TestAccountInsert)
	t.Run("testAccountQuery", TestQueryAccounts)
	t.Run("testAccountUpdate", TestUpdateAccounts)
}

func TestAccountInsert(t *testing.T) {
	t.Log("on inserting accounts")

	for _, acc := range accs {
		err = accdb.InsertAccount(&acc)
		assert.NoError(t, err)
	}
	t.Log("successfully inserted accounts")
}

func TestQueryAccounts(t *testing.T) {
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

func TestUpdateAccounts(t *testing.T) {
	accs[0].Name = "Makavelli"
	accs[1].Name = "Notorious BIG"

	for _, acc := range accs {
		err = accdb.UpdateAccount(&acc)
		assert.NoError(t, err)
	}
}

func TestDeleteAccounts(t *testing.T) {
	assert.NoError(t, accdb.DeleteAccount(accs[0].ID))
}
