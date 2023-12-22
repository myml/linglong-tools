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

type PushArgs struct {
	LayerFile   string
	RepoUrl     string
	RepoName    string
	RepoChannel string
	PrintStatus bool
}

var pushArgs = PushArgs{
	RepoName:    "repo",
	RepoChannel: "linglong",
}

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push linglong layer file to remote repository",
	Example: `# use environment variables: $LINGLONG_USERNAME and $LINGLONG_PASSOWRD (Recommend)
linglong-tools push -f ./test.layer -r https://repo.linglong.dev
# pass username and password
linglong-tools push -f ./test.layer -r https://user:pass@repo.linglong.dev
# pass repo name
linglong-tools push -f ./test.layer -r https://repo.linglong.dev -n develop-snipe`,
	Run: func(cmd *cobra.Command, args []string) {
		if pushArgs.PrintStatus {
			printStatus()
			return
		}
		err := pushRun(context.Background(), pushArgs)
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

func pushRun(ctx context.Context, args PushArgs) error {
	f, err := os.Open(args.LayerFile)
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
	u, err := url.Parse(args.RepoUrl)
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

	log.Println(UploadTaskStatusLogging)
	authReq := apiserver.RequestAuth{Username: &username, Password: &password}
	loginResp, _, err := api.ClientAPI.SignIn(ctx).Data(authReq).Execute()
	if err != nil {
		return fmt.Errorf("sign in: %w", err)
	}
	ref := fmt.Sprintf("%v/%v/%v/%v/%v",
		args.RepoChannel,
		info.Info.Appid,
		info.Info.Version,
		info.Info.Arch[0],
		info.Info.Module,
	)
	log.Println(UploadTaskStatusCreating)
	taskReq := apiserver.SchemaNewUploadTaskReq{RepoName: &args.RepoName, Ref: &ref}
	newTaskResp, _, err := api.ClientAPI.NewUploadTaskID(ctx).Req(taskReq).XToken(*loginResp.Data.Token).Execute()
	if err != nil {
		return fmt.Errorf("create upload task: %w", err)
	}
	log.Println(UploadTaskStatusUploading)
	_, _, err = api.ClientAPI.UploadTaskLayerFile(ctx, *newTaskResp.Data.Id).XToken(*loginResp.Data.Token).File(f).Execute()
	if err != nil {
		return fmt.Errorf("upload layer file: %w", err)
	}
	status := ""
	for {
		time.Sleep(time.Second)
		taskInfoResp, _, err := api.ClientAPI.UploadTaskInfo(ctx, *newTaskResp.Data.Id).XToken(*loginResp.Data.Token).Execute()
		if err != nil {
			return fmt.Errorf("get task info: %w", err)
		}
		latest := taskInfoResp.Data.GetStatus()
		if status != latest {
			status = latest
			log.Println(status)
		}
		if status == "failed" {
			return fmt.Errorf("task(%s) failed.", newTaskResp.Data.GetId())
		}
		if status == "complete" {
			return nil
		}
	}
}

func printStatus() {
	statusList := map[UploadTaskStatus]string{
		UploadTaskStatusLogging:   "",
		UploadTaskStatusCreating:  "",
		UploadTaskStatusUploading: "",
		UploadTaskStatusPending:   "",
		UploadTaskStatusUploaded:  "",
		UploadTaskStatusComplete:  "",
		UploadTaskStatusExtracted: "",
		UploadTaskStatusCommitted: "",
		UploadTaskStatusFailed:    "",
	}
	for i := range statusList {
		log.Println(statusList[i])
	}
}

type UploadTaskStatus string

var (
	/*** client status ***/
	UploadTaskStatusLogging   UploadTaskStatus = "logging"
	UploadTaskStatusCreating  UploadTaskStatus = "creating"
	UploadTaskStatusUploading UploadTaskStatus = "uploading"

	/*** server status ***/
	// UploadTaskStatusPending pending
	UploadTaskStatusPending UploadTaskStatus = "pending"
	// UploadTaskStatusUploaded uploaded
	UploadTaskStatusUploaded UploadTaskStatus = "uploaded"
	// UploadTaskStatusComplete complete
	UploadTaskStatusComplete UploadTaskStatus = "complete"
	// UploadTaskStatusExtracted extracted
	UploadTaskStatusExtracted UploadTaskStatus = "extracted"
	// UploadTaskStatusCommitted committed
	UploadTaskStatusCommitted UploadTaskStatus = "committed"
	// UploadTaskStatusFailed failed
	UploadTaskStatusFailed UploadTaskStatus = "failed"
)

func init() {
	pushCmd.Flags().StringVarP(&pushArgs.LayerFile, "file", "f", pushArgs.LayerFile, "layer file")
	pushCmd.Flags().StringVarP(&pushArgs.RepoUrl, "repo", "r", pushArgs.RepoUrl, "remote repo url")
	pushCmd.Flags().StringVarP(&pushArgs.RepoName, "name", "n", pushArgs.RepoName, "remote repo name")
	pushCmd.Flags().BoolVarP(&pushArgs.PrintStatus, "print", "p", pushArgs.PrintStatus, "print all status")
	pushCmd.Flags().StringVarP(&pushArgs.RepoChannel, "channel", "c", pushArgs.RepoChannel, "remote repo channel")
	err := pushCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(pushCmd)
}
