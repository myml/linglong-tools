package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/myml/linglong-tools/pkg/apiserver"
	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/spf13/cobra"
)

func initPushCmd() *cobra.Command {
	var pushArgs PushArgs
	pushCmd := cobra.Command{
		Use:   "push",
		Short: "Push linglong layer file to remote repository",
		Example: `  # use environment variables: $LINGLONG_USERNAME and $LINGLONG_PASSOWRD (Recommend)
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
			err := PushRun(context.Background(), pushArgs)
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
	pushCmd.Flags().StringVarP(&pushArgs.LayerFile, "file", "f", "", "layer file")
	pushCmd.Flags().StringVarP(&pushArgs.RepoUrl, "repo", "r", DefaultRepoUrl, "remote repo url")
	pushCmd.Flags().StringVarP(&pushArgs.RepoName, "name", "n", DefaultRepoName, "remote repo name")
	pushCmd.Flags().BoolVarP(&pushArgs.PrintStatus, "print", "p", false, "print all status")
	pushCmd.Flags().StringVarP(&pushArgs.RepoChannel, "channel", "c", DefaultChannel, "remote repo channel")
	err := pushCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	return &pushCmd
}

type PushArgs struct {
	LayerFile   string
	RepoUrl     string
	RepoName    string
	RepoChannel string
	PrintStatus bool
}

func PushRun(ctx context.Context, args PushArgs) error {
	f, err := os.Open(args.LayerFile)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	info, err := layer.ParseMetaInfo(f)
	if err != nil {
		return fmt.Errorf("parse layer info: %w", err)
	}
	log.Printf("%#v", info)
	_, err = f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("file seek: %w", err)
	}
	log.Println(UploadTaskStatusLogining)
	api, token, err := initAPIClient(args.RepoUrl, true)
	if err != nil {
		return fmt.Errorf("init api: %w", err)
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
	newTaskResp, _, err := api.NewUploadTaskID(ctx).Req(taskReq).XToken(*token).Execute()
	if err != nil {
		return fmt.Errorf("create upload task: %w", err)
	}
	if newTaskResp.GetData().Id == nil {
		return fmt.Errorf("task id is null: %s", *token)
	}
	log.Println(UploadTaskStatusUploading)
	_, _, err = api.UploadTaskLayerFile(ctx, *newTaskResp.Data.Id).XToken(*token).File(f).Execute()
	if err != nil {
		return fmt.Errorf("upload layer file: %w", err)
	}
	status := ""
	for {
		time.Sleep(time.Second)
		taskInfoResp, _, err := api.UploadTaskInfo(ctx, *newTaskResp.Data.Id).XToken(*token).Execute()
		if err != nil {
			return fmt.Errorf("get task info: %w", err)
		}
		if taskInfoResp.GetData().Status == nil {
			return fmt.Errorf("task status is null: %s", taskInfoResp.GetMsg())
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
	statusList := []struct {
		status UploadTaskStatus
		desc   string
	}{
		{UploadTaskStatusLogining, "正在登录"},
		{UploadTaskStatusCreating, "正在创建上传任务"},
		{UploadTaskStatusPending, "正在等待上传"},
		{UploadTaskStatusUploading, "正在上传文件"},
		{UploadTaskStatusUploaded, "上传文件完成"},
		{UploadTaskStatusExtracted, "已解压上传的文件"},
		{UploadTaskStatusCommitted, "已提交上传的文件"},
		{UploadTaskStatusComplete, "上传任务完成"},
		{UploadTaskStatusFailed, "上传任务失败"},
	}
	for _, status := range statusList {
		fmt.Println(status.status, "->", status.desc)
	}
}

type UploadTaskStatus string

var (
	/*** client status ***/
	UploadTaskStatusLogining  UploadTaskStatus = "logining"
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
