package dao_test

import (
	"crypto/sha512"
	b64 "encoding/base64"
	"encoding/binary"
	coredb "github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	corecfg "github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc/dao"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

var (
	fdb *dao.FileDB
	err error

	project1ID = coredb.NewID()
	org1ID     = coredb.NewID()
	account1ID = coredb.NewID()
)

func init() {
	var csc *corecfg.FileServiceConfig
	csc, err = corecfg.NewConfig("./testdata/db.yml")
	if err != nil {
		log.Fatalf("error initializing db: %v", err)
	}
	logger := log.New().WithField("test", "sys-file")
	logger.Level = log.DebugLevel
	testDb, err := coredb.NewCoreDB(logger, &csc.DBConfig, nil)
	if err != nil {
		log.Fatalf("error creating CoreDB: %v", err)
	}
	log.Debug("MakeSchema testing .....")
	fdb, err = dao.NewFileDB(testDb, logger)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully initialize sys-file-db:  %v", fdb)
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
	require.Equal(t, fileSum[:], avatarFile.Sum)

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
