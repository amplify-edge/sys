package mailer

import (
	"context"
	corepkg "github.com/getcouragenow/sys-share/sys-core/service/go/pkg"
	service "github.com/getcouragenow/sys/sys-core/service/go"
	"github.com/matcornic/hermes/v2"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

type MailSvc struct {
	senderName string
	senderMail string
	client     *sendgrid.Client
	logger     *logrus.Entry
	hp         hermes.Product
}

func NewMailSvc(mcfg *service.MailConfig, l *logrus.Entry) *MailSvc {
	return &MailSvc{
		senderName: mcfg.SenderName,
		senderMail: mcfg.SenderMail,
		client:     sendgrid.NewSendClient(mcfg.SendgridApiKey),
		logger:     l,
		hp: hermes.Product{
			Name:        mcfg.ProductName,
			Logo:        mcfg.LogoUrl,
			Copyright:   mcfg.Copyright,
			TroubleText: mcfg.TroubleContact,
		},
	}
}

func (m *MailSvc) GetHermesProduct() hermes.Product {
	return m.hp
}

func (m *MailSvc) SendMail(ctx context.Context, in *corepkg.EmailRequest) (*corepkg.EmailResponse, error) {
	sender := mail.NewEmail(m.senderName, m.senderMail)
	content := string(in.Content)
	for name, address := range in.Recipients {
		msg := mail.NewSingleEmail(sender, in.Subject, mail.NewEmail(name, address), content, content)
		resp, err := m.client.Send(msg)
		if err != nil {
			return &corepkg.EmailResponse{
				Success:        false,
				ErrMessage:     err.Error(),
				SuccessMessage: "",
			}, err
		}
		m.logger.Debugf("Email response: %s", resp.Body)
	}
	return &corepkg.EmailResponse{
		Success:        true,
		ErrMessage:     "",
		SuccessMessage: "Successfully sent all emails",
	}, nil
}
