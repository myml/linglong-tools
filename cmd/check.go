package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/myml/linglong-tools/pkg/uab"
	"github.com/spf13/cobra"
)

type CheckArgs struct {
	InputFile           string
	FormatOutput        string
	CheckSigned         bool
	CheckSystemdService bool
	// TODO
	CheckIcon    bool
	CheckDesktop bool
}

type CheckResult struct {
	Pass           bool
	Signed         *CheckResultItem `json:",omitempty"`
	SystemdService *CheckResultItem `json:",omitempty"`
}

type CheckResultItem struct {
	Pass    bool
	Message string
}

func initCheckCmd() *cobra.Command {
	var checkArgs = CheckArgs{}
	checkCmd := cobra.Command{
		Use:   "check",
		Short: "Package checker for linypas",
		Run: func(cmd *cobra.Command, args []string) {
			err := checkRun(checkArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	checkCmd.Flags().StringVarP(&checkArgs.InputFile, "file", "f", "", "input file")
	checkCmd.Flags().BoolVar(&checkArgs.CheckSystemdService, "systemd", true, "check systemd service files")
	checkCmd.Flags().BoolVar(&checkArgs.CheckSigned, "signed", true, "check is signed")
	checkCmd.Flags().StringVar(&checkArgs.FormatOutput, "format", "", "Format output using a custom template")
	err := checkCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	return &checkCmd
}

func checkRun(args CheckArgs) error {
	var result CheckResult

	dir, err := ioutil.TempDir("", "ll-check-")
	if err != nil {
		return fmt.Errorf("create temp dir failed: %w", err)
	}
	defer os.RemoveAll(dir)
	var in interface {
		HasSign() bool
		Extract(outputDir string) error
	}
	switch ext := filepath.Ext(args.InputFile); ext {
	case ".layer":
		layerFile, err := layer.NewLayer(args.InputFile)
		if err != nil {
			return fmt.Errorf("create layer from file failed: %w", err)
		}
		defer layerFile.Close()
		in = layerFile
	case ".uab":
		uabFile, err := uab.Open(args.InputFile)
		if err != nil {
			return fmt.Errorf("open uab file: %w", err)
		}
		defer uabFile.Close()
		in = uabFile
	default:
		return fmt.Errorf("file type %s is unsupported", args.InputFile)
	}
	// 检查是否签名
	if args.CheckSigned {
		result.Pass = false
		result.Signed = &CheckResultItem{Pass: in.HasSign()}
	}
	// 检查systemd service是否合规
	if args.CheckSystemdService {
		err = in.Extract(dir)
		if err != nil {
			return fmt.Errorf("extract file failed: %w", err)
		}
		cmd := exec.Command("bash")
		cmd.Dir = dir
		cmd.Stdin = bytes.NewReader([]byte(`cat info.json | grep kind | grep -v app && exit 0
 find files/lib/systemd | grep "/[a-z]*\.service$" && exit -1 || true`))
		out, err := cmd.CombinedOutput()
		if err != nil {
			result.Pass = false
			result.SystemdService = &CheckResultItem{Pass: false, Message: string(out)}
		} else {
			result.SystemdService = &CheckResultItem{Pass: true}
		}
	}
	// 自定义模板输出
	if len(args.FormatOutput) > 0 {
		tmpl, err := template.New("").Parse(args.FormatOutput)
		if err != nil {
			return fmt.Errorf("parse format: %w", err)
		}
		err = tmpl.Execute(os.Stdout, result)
		if err != nil {
			return fmt.Errorf("exec format: %w", err)
		}
		return nil
	} else {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}
		os.Stdout.Write(data)
	}
	return nil
}
