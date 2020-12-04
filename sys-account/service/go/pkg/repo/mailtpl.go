package repo

import (
	"fmt"
	"github.com/matcornic/hermes/v2"

	"github.com/getcouragenow/sys-share/sys-core/service/go/pkg/mailhelper"
)

func (ad *SysAccountRepo) mailVerifyAccountTpl(emailAddr string, verifyToken string, id string) ([]byte, error) {
	hb := hermes.Body{
		Name: emailAddr,
		Intros: []string{
			"Here's the verification code for your account",
		},
		Actions: []hermes.Action{
			{
				Instructions: "Put the following code to the verify account inputs",
				Button: hermes.Button{
					Color: "#22BC66", // Optional action button color
					Text:  "Input the verification code here",
					Link:  fmt.Sprintf("https://%s%s/%s", ad.domain, "sys-account/verify", id),
				},
				InviteCode: verifyToken,
			},
		},
		Title: "Verify Account",
	}
	email, err := mailhelper.ConstructEmail(ad.mail.GetHermesProduct(), hb)
	if err != nil {
		return nil, err
	}
	return email, nil
}

func (ad *SysAccountRepo) mailForgotPassword(emailAddr string, verifyToken string) ([]byte, error) {
	hb := hermes.Body{
		Name: emailAddr,
		Intros: []string{
			fmt.Sprintf("Hi %s, we received a reset password request on your account, to proceed please enter the token below", emailAddr),
		},
		Actions: []hermes.Action{
			{
				Instructions: "Put the following code to the verify account input",
				InviteCode:   verifyToken,
			},
		},
		Title: fmt.Sprintf("Reset Password Request for %s", emailAddr),
	}
	email, err := mailhelper.ConstructEmail(ad.mail.GetHermesProduct(), hb)
	if err != nil {
		return nil, err
	}
	return email, nil
}
