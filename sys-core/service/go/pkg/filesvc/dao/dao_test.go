package dao_test

import (
	"crypto/sha512"
	b64 "encoding/base64"
	"encoding/binary"
	"github.com/getcouragenow/sys-share/sys-core/service/logging/zaplog"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"

	utilities "github.com/getcouragenow/sys-share/sys-core/service/config"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	corecfg "github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc/dao"
)

var (
	fdb *dao.FileDB
	err error

	project1ID = utilities.NewID()
	org1ID     = utilities.NewID()
	account1ID = utilities.NewID()
)

func init() {
	logger := zaplog.NewZapLogger(zaplog.DEBUG, "test", true, "")
	logger.InitLogger(nil)
	var csc *corecfg.FileServiceConfig
	csc, err = corecfg.NewConfig("./testdata/db.yml")
	if err != nil {
		logger.Fatalf("error initializing db: %v", err)
	}
	testDb, err := coredb.NewCoreDB(logger, &csc.DBConfig, nil)
	if err != nil {
		logger.Fatalf("error creating CoreDB: %v", err)
	}
	logger.Debug("MakeSchema testing .....")
	fdb, err = dao.NewFileDB(testDb, logger)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("successfully initialize sys-file-db:  %v", fdb)
}

func TestAll(t *testing.T) {
	t.Run("Test Upsert File", testUpsertFile)
}

func testUpsertFile(t *testing.T) {
	f, err := ioutil.ReadFile("./testdata/59567750.png")
	require.NoError(t, err)
	destByte := b64.StdEncoding.EncodeToString(f)
	fileSum := sha512.Sum512(f)

	t.Log("inserting new file")
	daoDestByte, err := b64.StdEncoding.DecodeString(destByte)
	require.NoError(t, err)
	avatarFile, err := fdb.UpsertFromUploadRequest(daoDestByte, "", account1ID, false)
	require.NoError(t, err)
	require.Equal(t, fileSum[:], avatarFile.ShaHash)

	t.Log("upserting existing file")
	f, err = ioutil.ReadFile("./testdata/footer-gopher.jpg")
	require.NoError(t, err)
	destByte = b64.StdEncoding.EncodeToString(f)
	daoDestByte, err = b64.StdEncoding.DecodeString(destByte)
	require.NoError(t, err)
	avatarUpdated, err := fdb.UpsertFromUploadRequest(daoDestByte, "", account1ID, false)
	require.NoError(t, err)
	t.Logf("Binary size: %d", binary.Size(avatarUpdated.Binary))
	require.Equal(t, avatarUpdated.Id, avatarFile.Id)
	ioutil.WriteFile("./testdata/updatedAvatar.jpg", avatarUpdated.Binary, 0644)
}
