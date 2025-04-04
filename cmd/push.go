//go:build !disable_api
// +build !disable_api

package cmd

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/myml/linglong-tools/pkg/apiserver"
	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/myml/linglong-tools/pkg/tarutils"
	"github.com/myml/linglong-tools/pkg/types"
	"github.com/myml/linglong-tools/pkg/uab"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initPushCmd())
}

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
	pushCmd.Flags().StringVarP(&pushArgs.File, "file", "f", "", "layer file or tgz file")
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
	File        string
	RepoUrl     string
	RepoName    string
	RepoChannel string
	PrintStatus bool
}

// 推送tar流数据
func pushTarStream(ctx context.Context, args PushArgs, input io.Reader, ref string) error {
	log.Println(UploadTaskStatusLogining)
	api, token, err := initAPIClient(args.RepoUrl, true)
	if err != nil {
		return fmt.Errorf("init api: %w", err)
	}
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
	// openapi生成的上传文件代码需要将整个文件读取到内存，不适合大文件上传
	{
		r, w := io.Pipe()
		m := multipart.NewWriter(w)
		contentType := m.FormDataContentType()
		go func() {
			defer w.Close()
			part, err := m.CreateFormFile("file", "file.tgz")
			if err != nil {
				w.CloseWithError(err)
				return
			}
			if _, err = io.Copy(part, input); err != nil {
				w.CloseWithError(err)
				return
			}
			err = m.Close()
			if err != nil {
				w.CloseWithError(err)
				return
			}
		}()
		reqUrl := fmt.Sprintf("%s/api/v1/upload-tasks/%s/tar", args.RepoUrl, *newTaskResp.Data.Id)
		req, err := http.NewRequest(http.MethodPut, reqUrl, r)
		if err != nil {
			return fmt.Errorf("create http request: %w", err)
		}
		req.Header.Set("X-Token", *token)
		req.Header.Set("Content-Type", contentType)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("upload tgz file: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("upload tgz file: %s", resp.Status)
		}
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

// 推送layer目录
func pushDirectory(ctx context.Context, args PushArgs, directory string) error {
	data, err := os.ReadFile(filepath.Join(directory, "info.json"))
	if err != nil {
		return fmt.Errorf("read info.json: %w", err)
	}
	var info types.LayerInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return fmt.Errorf("unmarshal info.json: %w", err)
	}
	if info.ID == "" {
		info.ID = info.Appid
	}
	if info.Appid == "" {
		info.Appid = info.ID
	}
	ref := fmt.Sprintf("%v/%v/%v/%v/%v",
		args.RepoChannel,
		info.ID,
		info.Version,
		info.Arch[0],
		info.Module,
	)
	r, err := tarutils.CreateTarStream(directory)
	if err != nil {
		return fmt.Errorf("create tar stream: %w", err)
	}
	return pushTarStream(ctx, args, r, ref)
}

// 推送tgz文件
func pushTgz(ctx context.Context, args PushArgs) error {
	var info types.LayerInfo
	f, err := os.Open(args.File)
	if err != nil {
		return fmt.Errorf("open tgz file: %w", err)
	}
	defer f.Close()
	r, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip: file format error: %w", err)
	}
	defer r.Close()
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()
		if err != nil {
			return fmt.Errorf("tar: file format error: %w", err)
		}
		if header.Name == "./info.json" {
			err = json.NewDecoder(tarReader).Decode(&info)
			if err != nil {
				return fmt.Errorf("parse info.json: %w", err)
			}
			if len(info.ID) > 0 {
				info.Appid = info.ID
			}
			break
		}
	}
	ref := fmt.Sprintf("%v/%v/%v/%v/%v",
		args.RepoChannel,
		info.Appid,
		info.Version,
		info.Arch[0],
		info.Module,
	)
	_, err = f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("file seek: %w", err)
	}
	return pushTarStream(ctx, args, f, ref)
}

// 推送layer文件
func pushLayer(ctx context.Context, args PushArgs) error {
	l, err := layer.NewLayer(args.File)
	if err != nil {
		return fmt.Errorf("")
	}
	defer l.Close()
	// 签名文件需要将签名数据一起推送，所以先在本地解压
	if l.HasSign() {
		tmpDir, err := ioutil.TempDir("", "extract-layer-*")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)
		log.Println(UploadTaskStatusChecking)
		err = l.Extract(tmpDir)
		if err != nil {
			return fmt.Errorf("extract layer file: %w", err)
		}
		err = pushDirectory(ctx, args, tmpDir)
		if err != nil {
			return fmt.Errorf("push extracted files: %w", err)
		}
		return nil
	}

	f, err := os.Open(args.File)
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

	// openapi生成的上传文件代码需要将整个文件读取到内存，不适合大文件上传
	{
		f, err := os.Open(args.File)
		if err != nil {
			return fmt.Errorf("open tgz file: %w", err)
		}
		r, w := io.Pipe()
		m := multipart.NewWriter(w)
		contentType := m.FormDataContentType()
		go func() {
			defer w.Close()
			part, err := m.CreateFormFile("file", f.Name())
			if err != nil {
				w.CloseWithError(err)
				return
			}
			if _, err = io.Copy(part, f); err != nil {
				w.CloseWithError(err)
				return
			}
			err = m.Close()
			if err != nil {
				w.CloseWithError(err)
				return
			}
		}()
		reqUrl := fmt.Sprintf("%s/api/v1/upload-tasks/%s/layer", args.RepoUrl, *newTaskResp.Data.Id)
		req, err := http.NewRequest(http.MethodPut, reqUrl, r)
		if err != nil {
			return fmt.Errorf("create http request: %w", err)
		}
		req.Header.Set("X-Token", *token)
		req.Header.Set("Content-Type", contentType)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("upload layer file: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("upload layer file: %s", resp.Status)
		}
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

// 推送uab文件
func pushUab(ctx context.Context, args PushArgs) error {
	u, err := uab.Open(args.File)
	if err != nil {
		return fmt.Errorf("open uab file: %w", err)
	}
	defer u.Close()
	appLayerPath, err := u.AppLayerPath()
	if err != nil {
		return fmt.Errorf("get app layer path: %w", err)
	}
	tmpDir, err := ioutil.TempDir("", "extract-uab-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	log.Println(UploadTaskStatusChecking)
	err = u.Extract(tmpDir)
	if err != nil {
		return fmt.Errorf("extract uab file: %w", err)
	}
	err = pushDirectory(ctx, args, filepath.Join(tmpDir, appLayerPath))
	if err != nil {
		return fmt.Errorf("push extracted files: %w", err)
	}
	return nil
}

func PushRun(ctx context.Context, args PushArgs) error {
	switch ext := filepath.Ext(args.File); ext {
	case ".layer":
		return pushLayer(ctx, args)
	case ".tgz":
		return pushTgz(ctx, args)
	case ".uab":
		return pushUab(ctx, args)
	default:
		return fmt.Errorf("file type %s is unsupported", args.File)
	}
}

func printStatus() {
	statusList := []struct {
		status UploadTaskStatus
		desc   string
	}{
		{UploadTaskStatusChecking, "正在检查"},
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
	UploadTaskStatusChecking  UploadTaskStatus = "checking"
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
