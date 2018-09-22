package config

import (
	"errors"
	"fmt"

	"gopkg.in/go-playground/validator.v9"
)

// Valid monitor configuration value constants.
const (
	MonitorAll               = "*"
	MonitorNewProposals      = "new_proposals"
	MonitorActiveProposals   = "active_proposals"
	MonitorJailedValidators  = "jailed_validators"
	MonitorDoubleSigning     = "double_signing"
	MonitorMissingSignatures = "missing_signatures"
)

var (
	structValidate = validator.New()
	validValues    = map[string]struct{}{
		MonitorAll:               struct{}{},
		MonitorNewProposals:      struct{}{},
		MonitorActiveProposals:   struct{}{},
		MonitorJailedValidators:  struct{}{},
		MonitorDoubleSigning:     struct{}{},
		MonitorMissingSignatures: struct{}{},
	}
)

type (
	// Config defines the application's configuration structure.
	Config struct {
		PollInterval uint          `mapstructure:"poll_interval" validate:"required,gt=10"`
		Monitors     []string      `mapstructure:"monitors" validate:"required,validmonitor"`
		Database     Database      `mapstructure:"database" validate:"required,dive"`
		Targets      Targets       `mapstructure:"targets" validate:"required,dive"`
		Filters      Filters       `mapstructure:"filters" validate:"required,dive"`
		Network      NetworkConfig `mapstructure:"network" validate:"required,dive"`
		Integrations Integrations  `mapstructure:"integrations" validate:"required,dive"`
	}

	// Database defines embedded database configuration.
	Database struct {
		DataDir string `mapstructure:"data_dir" validate:"required"`
	}

	// NetworkConfig defines network related configuration.
	NetworkConfig struct {
		Clients []string `mapstructure:"clients" validate:"dive,url"`
	}

	// Targets defines alerting targets.
	Targets struct {
		Webhooks        []string `mapstructure:"webhooks" validate:"dive,url"`
		SMSRecipients   []string `mapstructure:"sms_recipients"`
		EmailRecipients []string `mapstructure:"email_recipients" validate:"dive,email"`
	}

	// Filters defines a set of validator address filters to match against when
	// monitoring and alerting.
	Filters struct {
		Validators []ValidatorFilter `mapstructure:"validator" validate:"required,dive"`
	}

	// ValidatorFilter defines a validator filter against.
	ValidatorFilter struct {
		Operator string `mapstructure:"operator" validate:"contains=cosmosaccaddr,required"`
		Address  string `mapstructure:"address" validate:"hexadecimal,required"`
	}

	// Integrations defines integration configuration for utilizing third-party
	// alerting tools.
	Integrations struct {
		SendGrid SendGridAPI `mapstructure:"sendgrid" validate:"required,dive"`
	}

	// SendGridAPI defines the required configuration for using the SendGrid API.
	SendGridAPI struct {
		Key      string `mapstructure:"api_key" validate:"required"`
		FromName string `mapstructure:"from_name" validate:"required"`
	}
)

func init() {
	structValidate.RegisterValidation("validmonitor", validateMonitor)
}

// validateMonitor implements the validator.Func interface. It validates if a
// valid series of monitor values were given.
func validateMonitor(fl validator.FieldLevel) bool {
	values := fl.Field().Interface().([]string)
	if len(values) == 0 {
		return false
	}

	for _, monitor := range values {
		if monitor == MonitorAll && len(values) > 1 {
			// cannot specify all and specific monitor values
			return false
		}

		if _, ok := validValues[monitor]; !ok {
			return false
		}
	}

	return true
}

// Validate performs basic validation of parsed application configuration. If
// any validation fails, an error is immediately returned.
func (cfg Config) Validate() error {
	if err := structValidate.Struct(cfg); err != nil {
		return newConfigErr(err)
	} else if len(cfg.Targets.EmailRecipients) == 0 &&
		len(cfg.Targets.SMSRecipients) == 0 &&
		len(cfg.Targets.Webhooks) == 0 {
		return newConfigErr(errors.New("no alert targets provided"))
	}

	return nil
}

func newConfigErr(err error) error {
	return fmt.Errorf("invalid configuration: \"%s\"", err)
}
