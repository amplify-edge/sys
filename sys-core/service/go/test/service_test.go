package db_test

import (
	"github.com/sirupsen/logrus"
	"testing"

	"github.com/stretchr/testify/assert"

	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/service"
)

const (
	addForeignIdx = "CREATE UNIQUE INDEX IF NOT EXISTS idx_some_datas_foreign_id ON some_datas(foreign_id)"
	tableName = "some_datas"
)

type SomeData struct {
	ID        string `genji:"id"`
	ForeignID string `genji:"foreign_id"`
	Blah      string `genji:"blah"`
}

func (s *SomeData) CreateSQL() []string {
	return coresvc.NewTable(tableName, map[string]string{
		"id":         "TEXT",
		"foreign_id": "TEXT",
		"blah":       "TEXT",
	}, []string{addForeignIdx}).CreateTable()
}

func insertSomeDatas(id, foreignId, blah string) error {
	return sysCoreSvc.Exec("INSERT INTO some_datas(id, foreign_id, blah) VALUES(?, ?, ?)", id, foreignId, blah)
}

func get(id string) (*SomeData, error) {
	var sd SomeData
	resp, err := sysCoreSvc.QueryOne("SELECT id, foreign_id, blah FROM some_datas WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	err = resp.StructScan(&sd)
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
	id := coresvc.NewID()
	foreignID := coresvc.NewID()
	err := insertSomeDatas(id, foreignID, "blahblah")
	assert.NoError(t, err)
	sd, err := get(id)
	assert.NoError(t,err)
	assert.Equal(t, id, sd.ID)
}
