//go:build !disable_api
// +build !disable_api

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/myml/linglong-tools/pkg/apiserver"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initSearchCmd())
}

type RemoteInfoArgs struct {
	RepoUrl     string
	RepoName    string
	RepoChannel string
	AppID       string
	Arch        string
	Version     string
	Module      string

	PrettierOutput bool
}

func RemoteInfoRun(ctx context.Context, args RemoteInfoArgs) error {
	client, _, err := initAPIClient(args.RepoUrl, false)
	if err != nil {
		return fmt.Errorf("init api client: %w", err)
	}
	result, _, err := client.SearchApp(ctx).
		RepoName(args.RepoName).
		Channel(args.RepoChannel).
		AppId(args.AppID).
		Module(args.Module).
		Arch(args.Arch).
		Version(args.Version).
		Execute()
	if err != nil {
		var apiError *apiserver.GenericOpenAPIError
		if errors.As(err, &apiError) {
			if json.Unmarshal(apiError.Body(), &result) == nil && result.GetCode() == 500 {
				os.Stdout.WriteString("[]\n")
				return nil
			}
		}
		return fmt.Errorf("send api request: %w", err)
	}
	// module 从 runtime 改成 binrary, 在这里做兼容
	if len(result.GetData()) == 0 && args.Module == DefaultModule {
		args.Module = "runtime"
		return RemoteInfoRun(ctx, args)
	}
	encoder := json.NewEncoder(os.Stdout)
	if args.PrettierOutput {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(result.GetData())
}

func initSearchCmd() *cobra.Command {
	var searchArgs RemoteInfoArgs
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search app info from remote repo",
		Example: `  # search for application with id: org.deepin.org, version: 1.5.4
  linglong-tools search -r https://repo.linglong.dev -i org.deepin.home -c main -v 1.5.4 -p`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RemoteInfoRun(context.Background(), searchArgs)
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
	searchCmd.Flags().StringVarP(&searchArgs.RepoUrl, "repo", "r", DefaultRepoUrl, "remote repo url")
	searchCmd.Flags().StringVarP(&searchArgs.RepoName, "name", "n", DefaultRepoName, "remote repo name")
	searchCmd.Flags().StringVarP(&searchArgs.RepoChannel, "channel", "c", DefaultChannel, "remote repo channel")
	searchCmd.Flags().StringVarP(&searchArgs.AppID, "app_id", "i", "", "app id")
	searchCmd.Flags().StringVarP(&searchArgs.Arch, "app_arch", "a", DefaultArch, "app arch")
	searchCmd.Flags().StringVarP(&searchArgs.Module, "app_module", "m", DefaultModule, "app module")
	searchCmd.Flags().StringVarP(&searchArgs.Version, "app_version", "v", "", "app version")
	searchCmd.Flags().BoolVarP(&searchArgs.PrettierOutput, "prettier", "p", false, "output pretty JSON")
	err := searchCmd.MarkFlagRequired("app_id")
	if err != nil {
		panic(err)
	}
	return searchCmd
}
