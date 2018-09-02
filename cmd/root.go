package cmd

import (
	"fmt"
	"os"

	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/version"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagConfig = "config"
)

// command flags
var (
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "titan",
	Short: "Titan is a configurable daemon that monitors and alerts validators in a Cosmos network",
	Long: `Titan is a configurable daemon that monitors and alerts validators when
critical and vital network events occur in a Cosmos network. Validators can run
Titan along side their nodes to stay up to date to the latest network events
only particular to them.`,
	Args: cobra.NoArgs,
	RunE: executeRootCmd,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVar(&configFile, flagConfig, "", "The daemon configuration file")
	// TODO: Define additional flags

	rootCmd.AddCommand(version.VersionCmd)
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// set the default configuration path in the user's home directory
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".titan")
		viper.SetConfigType("toml")
	}
}

// executeRootCmd implements the root command handler. It returns an error if
// the command failed to execute correctly.
func executeRootCmd(cmd *cobra.Command, args []string) error {
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	cfg := config.Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	// TODO: Database, RPC and logging

	return nil
}

// Execute executes the application root command. If any error is returned, it
// is printed to STDOUT and a non-zero exit status is returned.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
