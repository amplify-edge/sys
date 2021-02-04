// Provide readonly access to superuser config specified during creation
// Thus the API doesn't allow creation of new superuser from it, nor does it allow deletion.
package superusers

import (
	"context"
	"errors"
	"github.com/amplify-cms/sys-share/sys-account/service/go/pkg"
	sharedAuth "github.com/amplify-cms/sys-share/sys-account/service/go/pkg/shared"
	"io/ioutil"
	"strings"

	util "github.com/amplify-cms/sys-share/sys-core/service/config"
)

// toPkgAccount converts SuperUser struct to its proto counterparts.
func (s *SuperUser) toPkgAccount() *pkg.Account {
	return &pkg.Account{
		Id: s.Name,
		Role: []*pkg.UserRoles{
			{
				Role: pkg.SUPERADMIN,
			},
		},
		Email:    s.Name,
		Avatar:   []byte(s.Avatar),
		Password: s.HashedPassword,
		Verified: true,
	}
}

func (s *SuperUserIO) get(name string) (*SuperUser, error) {
	supes, err := s.readFile()
	if err != nil {
		return nil, err
	}
	for _, supe := range supes.SuperUsers {
		if supe.Name == name {
			return supe, nil
		}
	}
	return nil, errors.New("error: superuser not found")
}

func (s *SuperUserIO) readFile() (*SuperUserConfig, error) {
	bdata, err := ioutil.ReadFile(s.fpath)
	if err != nil {
		return nil, err
	}
	var supes SuperUserConfig
	err = util.UnmarshalYAML(bdata, &supes)
	if err != nil {
		return nil, err
	}
	return &supes, nil
}

func (s *SuperUserIO) verifyPermission(ctx context.Context) bool {
	claims := sharedAuth.ObtainClaimsFromContext(ctx)
	return sharedAuth.IsSuperadmin(claims.Role)
}

// Get Superuser based on its / his / her name.
// will return SuperUser object and error if any
func (s *SuperUserIO) Get(ctx context.Context, name string) (*pkg.Account, error) {
	//if !s.verifyPermission(ctx) {
	//	return nil, errors.New(sharedAuth.Error{Reason: sharedAuth.ErrInsufficientRights}.Error())
	//}
	supe, err := s.get(name)
	if err != nil {
		return nil, err
	}
	return supe.toPkgAccount(), nil
}

func (s *SuperUserIO) List(ctx context.Context, nameLike string) ([]*pkg.Account, error) {
	if !s.verifyPermission(ctx) {
		return nil, errors.New(sharedAuth.Error{Reason: sharedAuth.ErrInsufficientRights}.Error())
	}
	supes, err := s.readFile()
	if err != nil {
		return nil, err
	}
	var accounts []*pkg.Account
	if nameLike == "" {
		for _, supe := range supes.SuperUsers {
			if strings.Contains(supe.Name, nameLike) {
				accounts = append(accounts, supe.toPkgAccount())
			}
		}
		return accounts, nil
	}
	for _, supe := range supes.SuperUsers {
		accounts = append(accounts, supe.toPkgAccount())
	}
	return accounts, nil
}
