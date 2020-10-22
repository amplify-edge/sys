package db_test

import (
	"github.com/sirupsen/logrus"
	"testing"

	"github.com/stretchr/testify/assert"

	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/service"
)

type SomeData struct {
	ID        string `genji:"id"`
	ForeignID string `genji:"foreign_id"`
	Bleh      string `genji:"bleh"`
}

func (s *SomeData) CreateSQL() []string {
	return coresvc.NewTable("some_data", map[string]string{
		"id":         "TEXT",
		"foreign_id": "TEXT",
		"bleh":       "TEXT",
	}, []string{"CREATE UNIQUE INDEX IF NOT EXISTS idx_some_datas_foreign_id ON some_datas(foreign_id)"}).CreateTable()
}

func (s *SomeData) Insert(id, foreignId, bleh string) error {
	return sysCoreSvc.Exec("INSERT INTO some_datas VALUES(?, ?, ?)", id, foreignId, bleh)
}

func (s *SomeData) Get(id string) (*SomeData, error) {
	var sd SomeData
	resp, err := sysCoreSvc.QueryOne("SELECT id, foreign_id, bleh FROM some_datas WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	err = resp.StructScan(sd)
	if err != nil {
		return nil, err
	}
	return &sd, nil
}

func testCoreDBService(t *testing.T) {
	var err error
	logger := logrus.New().WithField("sys-db", "test")
	sysCoreSvc, err = coresvc.NewCoreDB(logger, sysCoreCfg)
	assert.NoError(t, err)

	t.Logf("sys-core-svc: %v", *sysCoreSvc)
}

func testTableCreation(t *testing.T) {
	err := sysCoreSvc.RegisterModels(map[string]coresvc.DbModel{
		"some_datas": &SomeData{},
	})
	assert.NoError(t, err)
	err = sysCoreSvc.MakeSchema()
	assert.NoError(t, err)
}

func testTableInsert(t *testing.T) {

}
