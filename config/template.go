package config

// NOTE: any changes here must be reflected in the config structure definition
const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

# Poll interval in seconds
poll_interval = 15

# Monitors to enable. All the valid monitors are listed below
#
# NOTE: '*' reflects enabling all monitors
monitors = [
  "new_proposals",
  "active_proposals",
  "jailed_validators",
  "double_signing",
  "missing_signatures",
]

# Data directory used for the embedded database
[database]
data_dir = "/Users/aleksbez/.titan/data"

# Network configuration including a list of trusted LCD endpoints
[network]
listen_addr = "0.0.0.0:36655"

# NOTE: These will be used in a round-robin fashion
clients = ["https://gaia-seeds.interblock.io:1317"]

# List of alerting targets
#
# NOTE: Webhooks are currently not supported and SMS and email targets are
# triggered via the same SendGrid API
[targets]
webhooks = []
sms_recipients = ["+11234567890"]
email_recipients = ["foo@bar.com"]

# A list of validator filters to filter against when executing monitors
#
# Note a validator operator must have a valid Bech32 prefix and the address must
# be a valid HEX address
[filters]
  [filters.validator]
    operator = "cosmosaccaddr1chchjxgackcqkn9fqgpsc4n9xamx4flgndapzg"
    address = "DBA70FA7E9D55E035AD87B41C4DC0C38511FD09A"

# A list of API integration configurations
#
# NOTE: Only SendGrid is supported at the moment
[integrations]
  [integrations.sendgrid]
    api_key = ""
    from_name = "Cosmos Titan"
`
