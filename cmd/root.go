package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "linglong-tools",
	Short: "A linglong tools. See https://github.com/myml/linglong-tools/README.md",
}

func Execute() {
	rootCmd.AddCommand(initInfoCmd())
	rootCmd.AddCommand(initPushCmd())
	rootCmd.AddCommand(initVersionCmd())
	rootCmd.AddCommand(initSearchCmd())
	rootCmd.AddCommand(initDeleteCmd())
	rootCmd.AddCommand(initExtractCmd())
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
