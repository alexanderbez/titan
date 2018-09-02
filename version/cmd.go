package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VersionCmd implements the version sub-command.
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the Titan daemon version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ClientVersion())
	},
}
