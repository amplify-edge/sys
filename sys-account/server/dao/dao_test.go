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

func TestAccountInsert(t *testing.T) {
	t.Log("on inserting accounts")
	accs := []dao.Account{
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
	for _, acc := range accs {
		err = accdb.InsertAccount(&acc)
		assert.NoError(t, err)
	}
	t.Log("successfully inserted accounts")
}


//
// func TestQuery(t *testing.T) {
// 	var o accounts.Org
// 	sql := fmt.Sprintf("SELECT * FROM " + o.TableName() + " WHERE name = 'org002';")
// 	if err := db.QueryTable(testDb, &o, sql, func(out interface{}) {
// 		log.Printf("org => %v", out.(*accounts.Org))
// 	}); err != nil {
// 		t.Error(err)
// 	}
//
// 	var p accounts.Project
// 	sql = fmt.Sprintf("SELECT * FROM " + p.TableName() + " WHERE name = 'proj002';")
// 	if err := db.QueryTable(testDb, &p, sql, func(out interface{}) {
// 		log.Printf("proj => %v", out.(*accounts.Project))
// 	}); err != nil {
// 		t.Error(err)
// 	}
//
// 	var u accounts.User
// 	sql = fmt.Sprintf("SELECT * FROM " + u.TableName() + " WHERE name = 'user2';")
// 	if err := db.QueryTable(testDb, &u, sql, func(out interface{}) {
// 		log.Printf("user => %v", out.(*accounts.User))
// 	}); err != nil {
// 		t.Error(err)
// 	}
//
// 	var r accounts.Roles
// 	sql = fmt.Sprintf("SELECT * FROM " + r.TableName() + " WHERE role = 'user';")
// 	if err := db.QueryTable(testDb, &r, sql, func(out interface{}) {
// 		log.Printf("role => %v", out.(*accounts.Roles))
// 	}); err != nil {
// 		t.Error(err)
// 	}
//
// 	var pr accounts.Permission
// 	sql = fmt.Sprintf("SELECT * FROM " + pr.TableName() + " WHERE org = 'org002' AND user = 'user2';")
// 	if err := db.QueryTable(testDb, &pr, sql, func(out interface{}) {
// 		log.Printf("promission => %v", out.(*accounts.Permission))
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }
//
// func TestDelete(t *testing.T) {
// 	log.Print("Clanup all tables .....")
// 	o := accounts.Org{}
// 	sql := "DELETE FROM " + o.TableName() + " WHERE name = 'org002';"
// 	log.Printf("DELETE Table: %v\n sql = %v", o.TableName(), sql)
// 	if err := testDb.Exec(sql); err != nil {
// 		t.Error(err)
// 	}
//
// 	u := accounts.User{}
// 	sql = "DELETE FROM " + u.TableName() + " WHERE name = 'user2';"
// 	log.Printf("DELETE Table: %v\n sql = %v", u.TableName(), sql)
// 	if err := testDb.Exec(sql); err != nil {
// 		t.Error(err)
// 	}
//
// 	p := accounts.Project{}
// 	sql = "DELETE FROM " + p.TableName() + " WHERE name = 'proj002';"
// 	log.Printf("DELETE Table: %v\n sql = %v", p.TableName(), sql)
// 	if err := testDb.Exec(sql); err != nil {
// 		t.Error(err)
// 	}
//
// 	r := accounts.Roles{}
// 	sql = "DELETE FROM " + r.TableName() + " WHERE role = 'user';"
// 	log.Printf("DELETE Table: %v\n sql = %v", r.TableName(), sql)
// 	if err := testDb.Exec(sql); err != nil {
// 		t.Error(err)
// 	}
//
// 	pr := accounts.Permission{}
// 	sql = "DELETE FROM " + pr.TableName() + " WHERE org = 'org002' AND user = 'user2';"
// 	log.Printf("DELETE Table: %v\n sql = %v", pr.TableName(), sql)
// 	if err := testDb.Exec(sql); err != nil {
// 		t.Error(err)
// 	}
// }
//
// func TestFinalResult(t *testing.T) {
// 	// If the final data is empty, it means that the test passed
// 	log.Print("Print result datas .....")
// 	printTables(t)
// }
//
// func printTables(t *testing.T) {
// 	var o accounts.Org
// 	sql := fmt.Sprintf("SELECT * FROM " + o.TableName() + ";")
// 	if err := db.QueryTable(testDb, &o, sql, func(out interface{}) {
// 		log.Printf("org => %v", out.(*accounts.Org))
// 	}); err != nil {
// 		t.Error(err)
// 	}
//
// 	var p accounts.Project
// 	sql = fmt.Sprintf("SELECT * FROM " + p.TableName() + ";")
// 	if err := db.QueryTable(testDb, &p, sql, func(out interface{}) {
// 		log.Printf("proj => %v", out.(*accounts.Project))
// 	}); err != nil {
// 		t.Error(err)
// 	}
//
// 	var u accounts.User
// 	sql = fmt.Sprintf("SELECT * FROM " + u.TableName() + ";")
// 	if err := db.QueryTable(testDb, &u, sql, func(out interface{}) {
// 		log.Printf("user => %v", out.(*accounts.User))
// 	}); err != nil {
// 		t.Error(err)
// 	}
//
// 	var r accounts.Roles
// 	sql = fmt.Sprintf("SELECT * FROM " + r.TableName() + ";")
// 	if err := db.QueryTable(testDb, &r, sql, func(out interface{}) {
// 		log.Printf("role => %v", out.(*accounts.Roles))
// 	}); err != nil {
// 		t.Error(err)
// 	}
//
// 	var pr accounts.Permission
// 	sql = fmt.Sprintf("SELECT * FROM " + pr.TableName() + ";")
// 	if err := db.QueryTable(testDb, &pr, sql, func(out interface{}) {
// 		log.Printf("promission => %v", out.(*accounts.Permission))
// 	}); err != nil {
// 		t.Error(err)
// 	}
// }
