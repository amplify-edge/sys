package delivery_test

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getcouragenow/sys-share/pkg"

	"github.com/getcouragenow/sys/sys-account/service/go/delivery"
	"github.com/getcouragenow/sys/sys-account/service/go/pkg/auth"
)

var (
	ad            *delivery.AuthDelivery
	loginRequests = []*pkg.LoginRequest{
		{
			Email:    "someemail@example.com",
			Password: "someInsecureBlaBlaPassword",
		},
		{
			Email:    "superadmin@getcouragenow.org",
			Password: "superadmin",
		},
	}
)

func TestAuthDeliveryAll(t *testing.T) {
	os.Setenv("JWT_ACCESS_SECRET", "AccessVerySecretHush!")
	os.Setenv("JWT_REFRESH_SECRET", "RefreshVeryHushHushFriends!")
	tc := auth.NewTokenConfig([]byte(os.Getenv("JWT_ACCESS_SECRET")), []byte(os.Getenv("JWT_REFRESH_SECRET")))
	ad = &delivery.AuthDelivery{
		Log:      logrus.New().WithField("test", "auth-delivery"),
		TokenCfg: tc,
	}
	t.Run("Test Login User", testUserLogin)
	t.Parallel()
}

func testUserLogin(t *testing.T) {
	// empty request
	_, err := ad.Login(context.Background(), nil)
	assert.Error(t, err, status.Errorf(codes.Unauthenticated, "Can't authenticate: %v", auth.AuthError{Reason: auth.ErrInvalidParameters}))
	// Wrong credentials
	_, err = ad.Login(context.Background(), loginRequests[0])
	assert.Error(t, err, status.Errorf(codes.Unauthenticated, "cannot authenticate: %v", auth.AuthError{Reason: auth.ErrInvalidCredentials}))
	// Correct Credentials
	resp, err := ad.Login(context.Background(), loginRequests[1])
	assert.NoError(t, err)
	t.Logf("Successfully logged in user: %s => %s, %s",
		loginRequests[1].Email, resp.AccessToken, resp.RefreshToken)
}
