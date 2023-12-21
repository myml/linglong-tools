/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/myml/linglong-tools/pkg/apiserver"
	"github.com/myml/linglong-tools/pkg/layer"
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
linglong-tools push -f ./test.layer -r https://user:pass@repo.linglong.dev
# pass repo name
linglong-tools push -f ./test.layer -r https://repo.linglong.dev -n develop-snipe`,
	Run: func(cmd *cobra.Command, args []string) {
		err := pushRun(context.Background())
		if err != nil {
			var apiError *apiserver.GenericOpenAPIError
			if errors.As(err, &apiError) {
				log.Fatalln(err, string(apiError.Body()))
			} else {
				log.Fatalln(err)
			}
		}
	},
}

func pushRun(ctx context.Context) error {
	f, err := os.Open(LayerFile)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	info, err := layer.ParseMetaInfo(f)
	if err != nil {
		return fmt.Errorf("parse layer info: %w", err)
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("file seek: %w", err)
	}
	u, err := url.Parse(RepoUrl)
	if err != nil {
		return fmt.Errorf("parse repo url: %w", err)
	}
	username, password := os.Getenv("LINGLONG_USERNAME"), os.Getenv("LINGLONG_PASSWORD")
	if u.User != nil {
		username = u.User.Username()
		pwd, ok := u.User.Password()
		if ok {
			password = pwd
		}
	}
	cfg := apiserver.NewConfiguration()
	cfg.Scheme = u.Scheme
	cfg.Host = u.Host
	api := apiserver.NewAPIClient(cfg)

	authReq := apiserver.RequestAuth{Username: &username, Password: &password}
	loginResp, _, err := api.ClientAPI.SignIn(ctx).Data(authReq).Execute()
	if err != nil {
		return fmt.Errorf("sign in: %w", err)
	}
	ref := fmt.Sprintf("%v/%v/%v/%v/%v", RepoChannel, info.Info.Appid, info.Info.Version, info.Info.Arch[0], info.Info.Module)
	taskReq := apiserver.SchemaNewUploadTaskReq{RepoName: &RepoName, Ref: &ref}
	newTaskResp, _, err := api.ClientAPI.NewUploadTaskID(ctx).Req(taskReq).XToken(*loginResp.Data.Token).Execute()
	if err != nil {
		return fmt.Errorf("create upload task: %w", err)
	}
	_, _, err = api.ClientAPI.UploadTaskLayerFile(ctx, *newTaskResp.Data.Id).XToken(*loginResp.Data.Token).File(f).Execute()
	if err != nil {
		return fmt.Errorf("upload layer file: %w", err)
	}
	for {
		time.Sleep(time.Second)
		taskInfoResp, _, err := api.ClientAPI.UploadTaskInfo(ctx, *newTaskResp.Data.Id).XToken(*loginResp.Data.Token).Execute()
		if err != nil {
			return fmt.Errorf("get task info: %w", err)
		}
		status := taskInfoResp.Data.GetStatus()
		if status == "failed" {
			return fmt.Errorf("task(%s) failed.", newTaskResp.Data.GetId())
		}
		log.Println("task status", status)
		if status == "complete" {
			return nil
		}
	}
}

func init() {
	pushCmd.Flags().StringVarP(&LayerFile, "file", "f", LayerFile, "layer file")
	pushCmd.Flags().StringVarP(&RepoUrl, "repo", "r", RepoUrl, "remote repo url")
	pushCmd.Flags().StringVarP(&RepoName, "name", "n", RepoName, "remote repo name")
	pushCmd.Flags().StringVarP(&RepoChannel, "channel", "c", RepoChannel, "remote repo channel")
	err := pushCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(pushCmd)
}
