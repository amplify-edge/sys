package repo

import (
	"context"
	"github.com/amplify-cms/sys-share/sys-core/service/logging"

	sharedAuth "github.com/amplify-cms/sys-share/sys-account/service/go/pkg/shared"
)

// Satisfies sharedAuth.ServerAuuthzInterceptor
func (ad *SysAccountRepo) GetUnauthenticatedRoutes() []string {
	return ad.unauthenticatedRoutes
}

func (ad *SysAccountRepo) GetLogger() logging.Logger {
	return ad.log
}

func (ad *SysAccountRepo) GetTokenConfig() *sharedAuth.TokenConfig {
	return ad.tokenCfg
}

func (ad *SysAccountRepo) GetAuthenticatedRoutes() map[string]func(claims sharedAuth.TokenClaims) error {
	return nil
}

// DefaultInterceptor is default authN/authZ interceptor, validates only token & claims correctness without performing any role specific authorization.
func (ad *SysAccountRepo) DefaultInterceptor(ctx context.Context) (context.Context, error) {
	return sharedAuth.ServerInterceptor(ad)(ctx)
}

// ObtainAccessClaimsFromMetadata obtains token claims from given context with gRPC metadata.
func (ad *SysAccountRepo) ObtainAccessClaimsFromMetadata(ctx context.Context, isAccess bool) (claims sharedAuth.TokenClaims, err error) {
	return sharedAuth.ClaimsFromMetadata(ctx, true, ad)
}
