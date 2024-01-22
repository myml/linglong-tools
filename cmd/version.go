package cmd

import (
	"fmt"

	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed version.txt
var Version string

func initVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(Version)
		},
	}
}
