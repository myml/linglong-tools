package cmd

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/stretchr/testify/require"
)

func TestPushRun(t *testing.T) {
	assert := require.New(t)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	var metaInfo layer.MetaInfo
	metaInfo.Info.Appid = "test"
	metaInfo.Info.Arch = append(metaInfo.Info.Arch, "amd64")
	fname := genLayerFile(assert, metaInfo)

	pushArgs.PrintStatus = true
	pushCmd.Run(nil, nil)
	// 模拟api服务
	fakeToken := "jwt_token_xxx"
	http.HandleFunc("/api/v1/sign-in", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// nolint
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]string{"token": fakeToken}})
	})
	http.HandleFunc("/api/v1/upload-tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") != fakeToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// nolint
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]string{"id": "upload_taskid_xxx"}})
	})
	http.HandleFunc("/api/v1/upload-tasks/upload_taskid_xxx/layer", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") != fakeToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	})
	http.HandleFunc("/api/v1/upload-tasks/upload_taskid_xxx/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Token") != fakeToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// nolint
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]string{"status": string(UploadTaskStatusComplete)}})
	})
	server := http.Server{Addr: ":8080", Handler: http.DefaultServeMux}
	go func() {
		pushArgs.LayerFile = fname
		pushArgs.RepoUrl = "http://test:pwd@127.0.0.1:8080"
		// 测试推送
		err := PushRun(context.Background(), pushArgs)
		assert.NoError(err)
		// 停止http服务
		err = server.Shutdown(context.Background())
		assert.NoError(err)
	}()
	// 开始http服务
	err := server.ListenAndServe()
	assert.ErrorIs(err, http.ErrServerClosed)
}
