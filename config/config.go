package config

import (
	"fmt"
)

type (
	// Config defines the application's configuration structure.
	Config struct {
		PollInterval uint           `mapstructure:"poll_interval"`
		Targets      *Targets       `mapstructure:"targets"`
		Filters      *Filters       `mapstructure:"filters"`
		Network      *NetworkConfig `mapstructure:"network"`
		Integrations *Integrations  `mapstructure:"integrations"`
	}

	// NetworkConfig defines network related configuration.
	NetworkConfig struct {
		LCDClients []string `mapstructure:"lcd_clients"`
	}

	// Targets defines alerting targets.
	Targets struct {
		Webhooks        []string `mapstructure:"webhooks"`
		SMSRecipients   []string `mapstructure:"sms_recipients"`
		EmailRecipients []string `mapstructure:"email_recipients"`
	}

	// Filters defines a set of validator address filters to match against when
	// monitoring and alerting.
	Filters struct {
		Validators []string `mapstructure:"validators"`
	}

	// Integrations defines integration configuration for utilizing third-party
	// alerting tools.
	Integrations struct {
		SendGrid *SendGridAPI `mapstructure:"sendgrid"`
	}

	// SendGridAPI defines the required configuration for using the SendGrid API.
	SendGridAPI struct {
		Key      string `mapstructure:"api_key"`
		FromName string `mapstructure:"from_name"`
	}
)

// Validate performs basic validation of parsed application configuration. If
// any validation fails, an error is immediately returned.
func (cfg Config) Validate() error {
	errPrefix := "invalid configuration"

	if len(cfg.Network.LCDClients) == 0 {
		return fmt.Errorf("%s: no LCD clients provided", errPrefix)
	}

	if cfg.Integrations.SendGrid.Key == "" {
		return fmt.Errorf("%s: no SendGrid API key provided", errPrefix)
	}

	if len(cfg.Targets.EmailRecipients) == 0 &&
		len(cfg.Targets.SMSRecipients) == 0 &&
		len(cfg.Targets.Webhooks) == 0 {
		return fmt.Errorf("%s: no alert targets provided", errPrefix)
	}

	return nil
}
