package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alexanderbez/titan/config"
)

func newTestValidConfig() config.Config {
	return config.Config{
		PollInterval: 15,
		Database:     config.Database{DataDir: "/test/data/dir"},
		Monitors:     []string{"*"},
		Targets: config.Targets{
			EmailRecipients: []string{"foo@bar.com"},
		},
		Filters: config.Filters{
			Validators: []config.ValidatorFilter{
				config.ValidatorFilter{
					Operator: "cosmosaccaddr1chchjxgackcqkn9fqgpsc4n9xamx4flgndapzg",
					Address:  "DBA70FA7E9D55E035AD87B41C4DC0C38511FD09A",
				},
			},
		},
		Network: config.NetworkConfig{
			Clients: []string{"https://test-seeds.com:1317"},
		},
		Integrations: config.Integrations{
			SendGrid: config.SendGridAPI{
				Key:      "test-key",
				FromName: "Cosmos Titan",
			},
		},
	}
}

func TestValidConfig(t *testing.T) {
	cfg := newTestValidConfig()

	err := cfg.Validate()
	require.NoError(t, err)
}

func TestEmptyConfig(t *testing.T) {
	cfg := config.Config{}

	err := cfg.Validate()
	require.Error(t, err)
}

func TestInvalidTargets(t *testing.T) {
	cfg := newTestValidConfig()
	cfg.Targets = config.Targets{}

	err := cfg.Validate()
	require.Error(t, err)
}

func TestInvalidMonitors(t *testing.T) {
	cfg := newTestValidConfig()

	cfg.Monitors = []string{}
	err := cfg.Validate()
	require.Error(t, err)

	cfg.Monitors = []string{"*", "new_proposals"}
	err = cfg.Validate()
	require.Error(t, err)

	cfg.Monitors = []string{"invalid_monitor", "new_proposals"}
	err = cfg.Validate()
	require.Error(t, err)
}
