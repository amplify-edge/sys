local tokencfg = import "vendor/github.com/getcouragenow/sys-share/sys-core/service/config/mixins/mixin.jwt.libsonnet";
{
    local cfg = self,
    UnauthenticatedRoutes:: [
    	"/v2.sys_account.services.AuthService/Login",
    	"/v2.sys_account.services.AuthService/Register",
    	"/v2.sys_account.services.AuthService/ResetPassword",
    	"/v2.sys_account.services.AuthService/ForgotPassword",
    	"/v2.sys_account.services.AuthService/RefreshAccessToken",
    	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
    ],
    AccessToken:: tokencfg.Token {
        secret: "some_jwt_access_secret",
        expiry: 3600,
    },
    RefreshToken:: tokencfg.Token {
        secret: "some_jwt_refresh_secret",
        expiry: cfg.AccessToken.expiry * 100,
    },
    sysAccountConfig: {
        unauthenticatedRoutes: cfg.UnauthenticatedRoutes,
        jwt: {
            access: cfg.AccessToken,
            refresh: cfg.RefreshToken,
        }
    }
}