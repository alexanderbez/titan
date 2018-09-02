package alerts

import (
	"fmt"

	"github.com/alexanderbez/titan/core"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var _ Sender = (*SendGridSender)(nil)

// SendGridSender implements a Sender via the SendGrid API. It is responsible
// for sending alerts to given recipient email and SMS addresses.
type SendGridSender struct {
	key         string
	fromAddress string
	fromName    string
	client      *sendgrid.Client
	logger      core.Logger
}

// NewSendGridSender returns a new SendGridSender.
func NewSendGridSender(logger core.Logger, key, fromName string) SendGridSender {
	return SendGridSender{
		key:         key,
		fromName:    fromName,
		fromAddress: "titan@sendgrid.net",
		client:      sendgrid.NewSendClient(key),
		logger:      logger,
	}
}

// Send implements the Sender interface. It will send an email (or SMS message)
// with a given payload (body) to a series of recipients. If any send fails, an
// error will be immediately returned.
//
// TODO: Investigate parallelizing sending messages.
func (sgs SendGridSender) Send(payload []byte, memo string, recipients []string) error {
	from := mail.NewEmail(sgs.fromName, sgs.fromAddress)
	subject := fmt.Sprintf("Titan Alert: %s", memo)

	for _, recipient := range recipients {
		to := mail.NewEmail("", recipient)
		message := mail.NewSingleEmail(from, subject, to, string(payload), "")

		response, err := sgs.client.Send(message)
		if err != nil {
			sgs.logger.Error(
				fmt.Sprintf("failed to send SendGrid alert; memo %s, recipient: %s, error: %v",
					memo, recipient, err,
				),
			)

			return err
		}

		sgs.logger.Debug(
			fmt.Sprintf("successfully sent SendGrid alert; memo %s, recipient: %s, response: %v",
				memo, recipient, response.Body,
			),
		)
	}

	return nil
}
