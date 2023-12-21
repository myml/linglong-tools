/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
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

var (
	FormatOutput   string
	PrettierOutput bool
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "get info of linglong layer file",
	Example: `linglong-tools info -f ./test.layer -p
linglong-tools info -f ./test.layer --format {{ .Info.Appid }}
linglong-tools info -f ./test.layer --format {{ .Raw }}`,
	Run: func(cmd *cobra.Command, args []string) {
		err := infoRun()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func infoRun() error {
	// 打开文件
	f, err := os.Open(LayerFile)
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
	if len(FormatOutput) > 0 {
		tmpl, err := template.New("").Parse(FormatOutput)
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
	if PrettierOutput {
		encoder.SetIndent("", "  ")
	}
	err = encoder.Encode(info)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	return nil
}

func init() {
	infoCmd.Flags().StringVarP(&LayerFile, "file", "f", LayerFile, "layer file")
	infoCmd.Flags().StringVar(&FormatOutput, "format", FormatOutput, "Format output using a custom template")
	infoCmd.Flags().BoolVarP(&PrettierOutput, "prettier", "p", PrettierOutput, "output pretty JSON")

	err := infoCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(infoCmd)
}
