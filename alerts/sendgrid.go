package alerts

import (
	"fmt"

	"github.com/alexanderbez/titan/core"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var _ Alerter = (*SendGridAlerter)(nil)

// SendGridAlerter implements an Alerter interface via the SendGrid API. It is
// responsible for sending alerts to given recipient email and SMS addresses.
type SendGridAlerter struct {
	fromAddress string
	fromName    string
	client      *sendgrid.Client
	logger      core.Logger
	recipients  []string
}

// NewSendGridAlerter returns a new SendGridAlerter.
func NewSendGridAlerter(logger core.Logger, apiKey, fromName string, recipients []string) SendGridAlerter {
	return SendGridAlerter{
		fromName:    fromName,
		fromAddress: "titan@sendgrid.net",
		client:      sendgrid.NewSendClient(apiKey),
		logger:      logger,
		recipients:  recipients,
	}
}

// Alert implements the Alerter interface. It will send an email (or SMS message)
// with a given payload (body) to a series of recipients. If any send fails, an
// error will be immediately returned.
//
// TODO: Investigate parallelizing sending messages.
func (sga SendGridAlerter) Alert(payload []byte, memo string) error {
	return sga.AlertWithRecipients(payload, memo, sga.recipients)
}

// AlertWithRecipients attempts to send a message to a series of recipients via
// the SendGrid API. If any send fails, an error will be immediately returned.
//
// TODO: Investigate parallelizing sending messages.
func (sga SendGridAlerter) AlertWithRecipients(payload []byte, memo string, recipients []string) error {
	from := mail.NewEmail(sga.fromName, sga.fromAddress)
	subject := fmt.Sprintf("Titan Alert: %s", memo)

	for _, recipient := range recipients {
		to := mail.NewEmail("", recipient)
		message := mail.NewSingleEmail(from, subject, to, string(payload), "")

		response, err := sga.client.Send(message)
		if err != nil {
			sga.logger.Error(
				fmt.Sprintf("failed to send SendGrid alert; memo %s, recipient: %s, error: %v",
					memo, recipient, err,
				),
			)

			return err
		}

		sga.logger.Debug(
			fmt.Sprintf("successfully sent SendGrid alert; memo %s, recipient: %s, response: %v",
				memo, recipient, response.Body,
			),
		)
	}

	return nil
}
