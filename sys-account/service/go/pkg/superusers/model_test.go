package superusers_test

import (
	"context"
	"github.com/amplify-cms/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/amplify-cms/sys-share/sys-account/service/go/pkg/shared"
	"github.com/amplify-cms/sys-share/sys-core/service/logging/zaplog"
	"github.com/amplify-cms/sys/sys-account/service/go/pkg/superusers"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	superIO       *superusers.SuperUserIO
	superFilePath = "./testdata/supers.yml"
)

func init() {
	logger := zaplog.NewZapLogger(zaplog.DEBUG, "superadmin-test", true, "")
	logger.InitLogger(nil)

	superIO = superusers.NewSuperUserDAO(superFilePath, logger)
}

func TestSuperUserDAO(t *testing.T) {
	t.Run("test initialization", testNewSuperuserDAO)
	t.Run("test get and list superuser", testGetAndListSuperuser)
}

func testNewSuperuserDAO(t *testing.T) {
	require.Equal(t, superIO.GetFilepath(), superFilePath)
}

func testGetAndListSuperuser(t *testing.T) {
	tcs := []sharedAuth.TokenClaims{
		{
			UserId:         superusers.DefaultSuperAdmin,
			Role:           nil,
			UserEmail:      superusers.DefaultSuperAdmin,
			StandardClaims: jwt.StandardClaims{},
		},
		{
			UserId: superusers.DefaultSuperAdmin,
			Role: []*pkg.UserRoles{
				{
					Role: pkg.SUPERADMIN,
				},
			},
			UserEmail:      superusers.DefaultSuperAdmin,
			StandardClaims: jwt.StandardClaims{},
		},
	}

	// should fail
	ctx := context.WithValue(context.Background(), sharedAuth.ContextKeyClaims, tcs[0])
	su, err := superIO.Get(ctx, "superadmin")
	require.Error(t, err)

	// should not fail
	ctx = context.WithValue(context.Background(), sharedAuth.ContextKeyClaims, tcs[1])
	su, err = superIO.Get(ctx, "superadmin")
	require.NoError(t, err)
	require.Equal(t, su.Id, "superadmin")
	require.Equal(t, []*pkg.UserRoles{
		{
			Role: pkg.SUPERADMIN,
		},
	}, su.Role)
}
