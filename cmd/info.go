package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/spf13/cobra"
)

func initInfoCmd() *cobra.Command {
	infoCmd := cobra.Command{
		Use:   "info",
		Short: "Get info of linglong layer file",
		Example: `  # output application information in json format
  linglong-tools info -f ./test.layer -p
  # Format output using a custom template (nesting)
  linglong-tools info -f ./test.layer --format '{{ .Info.Appid }}'
  # Format output using a custom template (array index)
  linglong-tools info -f ./test.layer --format '{{ index .Info.Arch 0 }}'`,
		Run: func(cmd *cobra.Command, args []string) {
			err := InfoRun(infoArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	infoCmd.Flags().StringVarP(&infoArgs.LayerFile, "file", "f", infoArgs.LayerFile, "layer file")
	infoCmd.Flags().StringVar(&infoArgs.FormatOutput, "format", infoArgs.FormatOutput, "Format output using a custom template")
	infoCmd.Flags().BoolVarP(&infoArgs.PrettierOutput, "prettier", "p", infoArgs.PrettierOutput, "output pretty JSON")
	err := infoCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	return &infoCmd
}

type InfoArgs struct {
	LayerFile      string
	FormatOutput   string
	PrettierOutput bool
}

var infoArgs = InfoArgs{}

// infoCmd represents the info command

func InfoRun(args InfoArgs) error {
	// 打开文件
	f, err := os.Open(args.LayerFile)
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
