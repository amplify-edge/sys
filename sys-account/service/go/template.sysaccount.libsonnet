{
    local cfg = self,
    UnauthenticatedRoutes:: [
    	"/v2.services.AuthService/Login",
    	"/v2.services.AuthService/Register",
    	"/v2.services.AuthService/ResetPassword",
    	"/v2.services.AuthService/ForgotPassword",
    	"/v2.services.AuthService/RefreshAccessToken",
    	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
    ],
    AccessToken:: {
        secret: "some_jwt_access_secret",
        expiry: 3600,
    },
    RefreshToken:: {
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