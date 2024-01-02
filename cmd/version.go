package cmd

import (
	"fmt"

	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed version.txt
var Version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
