# Fuzzy Search Command Implementation Plan

> **For agentic workers:** This is a single-task plan.

**Goal:** Add `linglong-tools fuzzy <keyword>` command for fuzzy searching apps.

**Architecture:** New cobra subcommand in `cmd/fuzzy.go`, modeled after existing `cmd/search.go`. Calls `POST /api/v0/apps/fuzzysearchapp` with keyword as `appId`.

**Tech Stack:** Go, cobra, openapi-generated client

## Global Constraints

- Build tag: `//go:build !disable_api`
- Must compile without errors
- Follow existing `search.go` code patterns exactly

---

### Task 1: Create `cmd/fuzzy.go`

**Files:**
- Create: `cmd/fuzzy.go`

**Interfaces:**
- Consumes: `apiserver.FuzzySearchApp`, `apiserver.RequestFuzzySearchReq`, `initAPIClient()` from `cmd/utils.go`
- Produces: Registers `fuzzy` subcommand on `rootCmd` via `init()`

- [ ] **Step 1: Create `cmd/fuzzy.go`**

```go
//go:build !disable_api
// +build !disable_api

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/myml/linglong-tools/v2/internal/apiserver"
	"github.com/spf13/cobra"
)

type FuzzySearchArgs struct {
	Keyword string

	RepoUrl     string
	RepoName    string
	RepoChannel string
	Arch        string
	Version     string

	PrettierOutput bool
}

func FuzzySearchRun(ctx context.Context, args FuzzySearchArgs) error {
	client, _, err := initAPIClient(args.RepoUrl, false)
	if err != nil {
		return fmt.Errorf("init api client: %w", err)
	}
	req := apiserver.NewRequestFuzzySearchReq()
	req.SetAppId(args.Keyword)
	req.SetRepoName(args.RepoName)
	req.SetChannel(args.RepoChannel)
	req.SetArch(args.Arch)
	if args.Version != "" {
		req.SetVersion(args.Version)
	}
	result, _, err := client.FuzzySearchApp(ctx).Data(*req).Execute()
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
	encoder := json.NewEncoder(os.Stdout)
	if args.PrettierOutput {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(result.GetData())
}

func init() {
	rootCmd.AddCommand(initFuzzySearchCmd())
}

func initFuzzySearchCmd() *cobra.Command {
	var fuzzyArgs FuzzySearchArgs
	fuzzyCmd := &cobra.Command{
		Use:   "fuzzy <keyword>",
		Short: "Fuzzy search apps from remote repo by keyword",
		Args:  cobra.ExactArgs(1),
		Example: `  # search for apps matching keyword
  linglong-tools fuzzy wechat
  linglong-tools fuzzy org.deepin.home -c main -a x86_64 -p`,
		Run: func(cmd *cobra.Command, args []string) {
			fuzzyArgs.Keyword = args[0]
			err := FuzzySearchRun(context.Background(), fuzzyArgs)
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
	fuzzyCmd.Flags().StringVarP(&fuzzyArgs.RepoUrl, "repo", "r", DefaultRepoUrl, "remote repo url")
	fuzzyCmd.Flags().StringVarP(&fuzzyArgs.RepoName, "name", "n", DefaultRepoName, "remote repo name")
	fuzzyCmd.Flags().StringVarP(&fuzzyArgs.RepoChannel, "channel", "c", DefaultChannel, "remote repo channel")
	fuzzyCmd.Flags().StringVarP(&fuzzyArgs.Arch, "app_arch", "a", DefaultArch, "app arch")
	fuzzyCmd.Flags().StringVarP(&fuzzyArgs.Version, "app_version", "v", "", "app version")
	fuzzyCmd.Flags().BoolVarP(&fuzzyArgs.PrettierOutput, "prettier", "p", false, "output pretty JSON")
	return fuzzyCmd
}
```

- [ ] **Step 2: Build and verify compilation**

Run: `go build ./...`
Expected: No errors, binary `linglong-tools` produced.

- [ ] **Step 3: Verify command is registered**

Run: `go run . fuzzy --help`
Expected: Shows help text for `fuzzy` subcommand.