package repo

import (
	"context"
	rpc "go.amplifyedge.org/sys-share-v2/sys-account/service/go/rpc/v2"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/logging/zaplog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sharedAuth "go.amplifyedge.org/sys-share-v2/sys-account/service/go/pkg/shared"
)

var (
	ad            *SysAccountRepo
	loginRequests = []*rpc.LoginRequest{
		{
			Email:    "someemail@example.com",
			Password: "someInsecureBlaBlaPassword",
		},
		{
			Email:    "superadmin@amplify-cms.org",
			Password: "superadmin",
		},
	}
)

func TestSysAccountRepoAll(t *testing.T) {
	os.Setenv("JWT_ACCESS_SECRET", "AccessVerySecretHush!")
	os.Setenv("JWT_REFRESH_SECRET", "RefreshVeryHushHushFriends!")
	tc := sharedAuth.NewTokenConfig([]byte(os.Getenv("JWT_ACCESS_SECRET")), []byte(os.Getenv("JWT_REFRESH_SECRET")))
	logger := zaplog.NewZapLogger(zaplog.DEBUG, "sys-account-repo-test", true, "")
	logger.InitLogger(nil)
	ad = &SysAccountRepo{
		log:      logger,
		tokenCfg: tc,
	}
	t.Run("Test Login User", testUserLogin)
	t.Parallel()
}

func testUserLogin(t *testing.T) {
	// empty request
	_, err := ad.Login(context.Background(), nil)
	assert.Error(t, err, status.Errorf(codes.Unauthenticated, "Can't sharedAuthenticate: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidParameters}))
	// Wrong credentials
	_, err = ad.Login(context.Background(), loginRequests[0])
	assert.Error(t, err, status.Errorf(codes.Unauthenticated, "cannot sharedAuthenticate: %v", sharedAuth.Error{Reason: sharedAuth.ErrInvalidCredentials}))
	// Correct Credentials
	resp, err := ad.Login(context.Background(), loginRequests[1])
	assert.NoError(t, err)
	t.Logf("Successfully logged in user: %s => %s, %s",
		loginRequests[1].Email, resp.AccessToken, resp.RefreshToken)
}
