package delivery

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getcouragenow/sys-share/pkg"
)

func (ad *AuthDelivery) GetAccount(ctx context.Context, in *pkg.GetAccountRequest) (*pkg.Account, error) {
	if in == nil {
		return &pkg.Account{}, status.Errorf(codes.InvalidArgument, "cannot get user account: %v", auth.AuthError{Reason: auth.ErrInvalidParameters})
	}

	return &pkg.Account{
		Id:       "1hpR8BL89uYI1ibPNgcRHI9Nn5Wi",
		Email:    "superadmin@getcouragenow.org",
		Password: "superadmin",
		Role: &pkg.UserRoles{
			Role: pkg.Roles(0),
			All:  true,
		},
	}, nil
}
