//go:build !disable_api
// +build !disable_api

package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/myml/linglong-tools/internal/apiserver"
)

const (
	DefaultRepoUrl  = "https://repo.linglong.dev"
	DefaultRepoName = "stable"
	DefaultChannel  = "main"
	DefaultModule   = "binary"
	DefaultArch     = "x86_64"
)

// 初始化API，如果login为false，token返回空
func initAPIClient(repoUrl string, login bool) (client *apiserver.ClientAPIService, token *string, err error) {
	u, err := url.Parse(repoUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("parse repo url: %w", err)
	}

	cfg := apiserver.NewConfiguration()
	cfg.Servers[0].URL = repoUrl
	api := apiserver.NewAPIClient(cfg)
	if login {
		username, password := os.Getenv("LINGLONG_USERNAME"), os.Getenv("LINGLONG_PASSWORD")
		if u.User != nil {
			username = u.User.Username()
			password, _ = u.User.Password()
		}
		authReq := apiserver.RequestAuth{Username: &username, Password: &password}
		loginResp, _, err := api.ClientAPI.SignIn(context.Background()).Data(authReq).Execute()
		if err != nil {
			return nil, nil, fmt.Errorf("sign in: %w", err)
		}
		if loginResp.GetData().Token == nil {
			return nil, nil, fmt.Errorf("token is null: %s", loginResp.GetMsg())
		}
		return api.ClientAPI, loginResp.Data.Token, nil
	}
	return api.ClientAPI, nil, nil
}
