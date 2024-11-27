package cmd

import (
	"fmt"

	_ "embed"

	"github.com/spf13/cobra"
)

func initVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(VERSION)
		},
	}
}
