package service_test

import (
	"fmt"
	commonCfg "github.com/getcouragenow/sys-share/sys-core/service/config/common"
	"testing"

	"github.com/stretchr/testify/assert"

	acccfg "github.com/getcouragenow/sys/sys-account/service/go"
)

func TestNewSysAccountConfig(t *testing.T) {
	baseTestDir := "./test/config"
	// Test nonexistent config
	_, err := acccfg.NewConfig("./nonexistent.yml")
	assert.Error(t, err)
	// Test valid config
	sysAccCfg, err := acccfg.NewConfig(fmt.Sprintf("%s/%s", baseTestDir, "valid.yml"))
	assert.NoError(t, err)
	expected := &acccfg.SysAccountConfig{
		SysAccountConfig: acccfg.Config{
			UnauthenticatedRoutes: []string{
				"/v2.AuthService/Login",
				"/v2.AuthService/Register",
				"/v2.AuthService/ResetPassword",
				"/v2.AuthService/ForgotPassword",
				"/v2.AuthService/RefreshAccessToken",
				"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
			},
			JWTConfig: acccfg.JWTConfig{
				Access: commonCfg.TokenConfig{
					Secret: "some_jwt_access_secret",
					Expiry: 3600,
				},
				Refresh: commonCfg.TokenConfig{
					Secret: "some_jwt_refresh_secret",
					Expiry: 3600,
				},
			},
		},
	}

	assert.Equal(t, expected, sysAccCfg)
}
