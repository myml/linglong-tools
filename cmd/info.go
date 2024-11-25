package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/myml/linglong-tools/pkg/uab"
	"github.com/spf13/cobra"
)

func initInfoCmd() *cobra.Command {
	infoCmd := cobra.Command{
		Use:   "info",
		Short: "Get info of linglong layer/uab file",
		Example: `  # output application information in json format
  linglong-tools info -f ./test.layer -p
  # Format output using a custom template (nesting)
  linglong-tools info -f ./test.layer --format '{{ .Info.Appid }}'
  # Format output using a custom template (array index)
  linglong-tools info -f ./test.layer --format '{{ index .Info.Arch 0 }}

  Operations of UAB file are the same as that of layer file`,
		Run: func(cmd *cobra.Command, args []string) {
			err := InfoRun(infoArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	infoCmd.Flags().StringVarP(&infoArgs.InputFile, "file", "f", infoArgs.InputFile, "input file")
	infoCmd.Flags().StringVar(&infoArgs.FormatOutput, "format", infoArgs.FormatOutput, "Format output using a custom template")
	infoCmd.Flags().BoolVarP(&infoArgs.PrettierOutput, "prettier", "p", infoArgs.PrettierOutput, "output pretty JSON")
	err := infoCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	return &infoCmd
}

type InfoArgs struct {
	InputFile      string
	FormatOutput   string
	PrettierOutput bool
}

var infoArgs = InfoArgs{}

// infoCmd represents the info command

func InfoRun(args InfoArgs) error {
	switch ext := filepath.Ext(args.InputFile); ext {
	case ".layer":
		return layerInfo(args)
	case ".uab":
		return uabInfo(args)
	default:
		return fmt.Errorf("file type %s is unsupported", args.InputFile)
	}
}

func layerInfo(args InfoArgs) error {
	// 打开文件
	f, err := os.Open(args.InputFile)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	// 读取信息
	info, err := layer.ParseMetaInfo(f)
	if err != nil {
		return fmt.Errorf("parse info: %w", err)
	}
	// 自定义模板输出
	if len(args.FormatOutput) > 0 {
		tmpl, err := template.New("").Parse(args.FormatOutput)
		if err != nil {
			return fmt.Errorf("parse format: %w", err)
		}
		err = tmpl.Execute(os.Stdout, info)
		if err != nil {
			return fmt.Errorf("exec format: %w", err)
		}
		return nil
	}
	encoder := json.NewEncoder(os.Stdout)
	if args.PrettierOutput {
		encoder.SetIndent("", "  ")
	}
	err = encoder.Encode(info)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	return nil
}

func uabInfo(args InfoArgs) error {
	uab, err := uab.Open(infoArgs.InputFile)
	if err != nil {
		return fmt.Errorf("open uab file: %w", err)
	}
	defer uab.Close()

	info := uab.MetaInfo()
	if len(args.FormatOutput) > 0 {
		template, err := template.New("").Parse(args.FormatOutput)
		if err != nil {
			return fmt.Errorf("parse format: %w", err)
		}

		err = template.Execute(os.Stdout, info)
		if err != nil {
			return fmt.Errorf("exec format: %w", err)
		}

		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	if args.PrettierOutput {
		encoder.SetIndent("", "  ")
	}
	err = encoder.Encode(info)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	return nil
}
