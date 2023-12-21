/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	RepoUrl     string
	RepoName    = "repo"
	RepoChannel = "linglong"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push linglong layer file to remote repository",
	Example: `# use environment variables: $LINGLONG_USERNAME and $LINGLONG_PASSOWRD (Recommend)
linglong-tools push -f ./test.layer -r https://repo.linglong.dev
# pass username and password
linglong-tools push -f ./test.layer -r https://user:pass@repo.linglong.dev ()
# pass repo name
linglong-tools push -f ./test.layer -r https://repo.linglong.dev -n develop-snipe`,
	Run: func(cmd *cobra.Command, args []string) {
		err := pushRun()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func pushRun() error {
	return fmt.Errorf("Not implemented")
}

func init() {
	pushCmd.Flags().StringVarP(&RepoUrl, "repo", "r", RepoUrl, "remote repo url")
	pushCmd.Flags().StringVarP(&RepoName, "name", "n", RepoName, "remote repo name")
	pushCmd.Flags().StringVarP(&RepoChannel, "channel", "c", RepoChannel, "remote repo channel")
	rootCmd.AddCommand(pushCmd)
}
