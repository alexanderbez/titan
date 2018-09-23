package alerts

import (
	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
)

// Alerter is an interface that defines a generic alerting hook.
type Alerter interface {
	Alert(payload []byte, memo string) error
	Name() string
}

// CreateAlerters creates the core series of alerting components used to alert
// a client with a monitor triggers an event.
func CreateAlerters(cfg config.Config, logger core.Logger) []Alerter {
	var sgRecipients []string
	sgRecipients = append(sgRecipients, cfg.Targets.EmailRecipients...)
	sgRecipients = append(sgRecipients, cfg.Targets.SMSRecipients...)

	sgAlerter := NewSendGridAlerter(
		logger.With("module", "SendGrid"),
		cfg.Integrations.SendGrid.Key,
		cfg.Integrations.SendGrid.FromName,
		sgRecipients,
	)

	return []Alerter{sgAlerter}
}
