package db_test

import (
	corecfg "github.com/getcouragenow/sys/sys-core/service/go"
	coresvc "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"testing"
)

var (
	sysCoreCfg *corecfg.SysCoreConfig
	sysCoreSvc *coresvc.CoreDB
)

func TestDBService(t *testing.T) {
	t.Run("Test Config Creation", testNewSysCoreConfig)
	t.Run("Test Service Creation", testCoreDBService)
	t.Run("Test Table Creation", testTableCreation)
	t.Run("Test Table Insertion", testTableInsert)
}
