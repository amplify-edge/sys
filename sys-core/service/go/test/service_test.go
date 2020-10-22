package db_test

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

const (
	tableName = "some_datas"
)

var (
	addForeignIdx = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_foreign_id ON %s(foreign_id)", tableName, tableName)
)

type SomeData struct {
	ID        string `genji:"id" validate:"required,len=27"`
	ForeignID string `genji:"foreign_id" validate:"required,len=27"`
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
	return sysCoreSvc.Exec(
		fmt.Sprintf("INSERT INTO %s(id, foreign_id, blah) VALUES(?, ?, ?)", tableName), id, foreignId, blah)
}

func get(id string) (*SomeData, error) {
	var sd SomeData
	resp, err := sysCoreSvc.QueryOne(
		fmt.Sprintf("SELECT id, foreign_id, blah FROM %s WHERE id = ?", tableName), id,
	)
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
	assert.NoError(t, err)

	validate := validator.New()

	err = validate.Struct(sd)
	assert.NoError(t, err)
	assert.Equal(t, id, sd.ID)
	assert.Equal(t, foreignID, sd.ForeignID)

}
