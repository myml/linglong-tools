/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/myml/linglong-tools/pkg/apiserver"
	"github.com/spf13/cobra"
)

type DeleteArgs struct {
	RepoUrl     string
	RepoName    string
	RepoChannel string
	AppID       string
	Arch        string
	Version     string
	Module      string
}

func DeleteRun(ctx context.Context, args DeleteArgs) error {
	client, token, err := initAPIClient(args.RepoUrl, true)
	if err != nil {
		return fmt.Errorf("init api client: %w", err)
	}
	_, err1 := client.RefDelete(ctx,
		args.RepoName,
		args.RepoChannel,
		args.AppID,
		args.Version,
		args.Arch,
		args.Module,
	).
		XToken(*token).
		Hard("true").
		Execute()

	_, err2 := client.RefDelete(ctx,
		args.RepoName,
		args.RepoChannel,
		args.AppID,
		args.Version,
		args.Arch,
		"runtime",
	).
		XToken(*token).
		Hard("true").
		Execute()
	// 只要有一个成功的就算成功
	if err1 == nil || err2 == nil {
		return nil
	}
	if err1 != nil {
		return fmt.Errorf("send api request: %w", err1)
	}
	return fmt.Errorf("send api request: %w", err2)
}

func initDeleteCmd() *cobra.Command {
	var cmdArgs DeleteArgs
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete app from remote repo",
		Example: `  # delete for application with id: org.deepin.org, version: 1.5.4
  linglong-tools delete -r https://repo.linglong.dev -i org.deepin.home -c main -v 1.5.4`,
		Run: func(cmd *cobra.Command, args []string) {
			err := DeleteRun(context.Background(), cmdArgs)
			if err != nil {
				var apiError *apiserver.GenericOpenAPIError
				if errors.As(err, &apiError) {
					log.Fatalln(err, "body: ", string(apiError.Body()))
				} else {
					log.Fatalln(err)
				}
			}
		},
	}
	cmd.Flags().StringVarP(&cmdArgs.RepoUrl, "repo", "r", DefaultRepoUrl, "remote repo url")
	cmd.Flags().StringVarP(&cmdArgs.RepoName, "name", "n", DefaultRepoName, "remote repo name")
	cmd.Flags().StringVarP(&cmdArgs.RepoChannel, "channel", "c", DefaultChannel, "remote repo channel")
	cmd.Flags().StringVarP(&cmdArgs.AppID, "app_id", "i", "", "app id")
	cmd.Flags().StringVarP(&cmdArgs.Arch, "app_arch", "a", DefaultArch, "app arch")
	cmd.Flags().StringVarP(&cmdArgs.Module, "app_module", "m", DefaultModule, "app module")
	cmd.Flags().StringVarP(&cmdArgs.Version, "app_version", "v", "", "app version")
	err := cmd.MarkFlagRequired("app_id")
	if err != nil {
		panic(err)
	}
	err = cmd.MarkFlagRequired("app_version")
	if err != nil {
		panic(err)
	}
	return cmd
}
