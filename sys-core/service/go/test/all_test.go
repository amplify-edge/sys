package db_test

import (
	corecfg "go.amplifyedge.org/sys-v2/sys-core/service/go"
	coresvc "go.amplifyedge.org/sys-v2/sys-core/service/go/pkg/coredb"
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
