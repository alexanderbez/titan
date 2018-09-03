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
	if len(cfg.Network.LCDClients) == 0 {
		return newConfigErr("no LCD clients provided")
	}

	if cfg.Integrations.SendGrid.Key == "" {
		return newConfigErr("no SendGrid API key provided")
	}

	if len(cfg.Targets.EmailRecipients) == 0 &&
		len(cfg.Targets.SMSRecipients) == 0 &&
		len(cfg.Targets.Webhooks) == 0 {
		return newConfigErr("no alert targets provided")
	}

	return nil
}

func newConfigErr(errStr string) error {
	return fmt.Errorf("invalid configuration: %s", errStr)
}
