package cmd

import (
	"debug/elf"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/spf13/cobra"
)

type MetaArgs struct {
	InputFile string
}

func initMetaCmd() *cobra.Command {
	var metaArgs MetaArgs
	metaCmd := cobra.Command{
		Use:   "meta",
		Short: "Get raw meta of linglong layer/uab file",
		Example: `  # output file meta
  linglong-tools meta -f ./test.layer
  linglong-tools meta -f ./test.uab
  You should try the info command first.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := MetaRun(metaArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	metaCmd.Flags().StringVarP(&metaArgs.InputFile, "file", "f", metaArgs.InputFile, "input file")
	err := metaCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	return &metaCmd
}

func MetaRun(args MetaArgs) error { // 打开文件
	switch ext := filepath.Ext(args.InputFile); ext {
	case ".layer":
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
		os.Stdout.Write([]byte(info.Raw))
	case ".uab":
		bin, err := elf.Open(args.InputFile)
		if err != nil {
			return fmt.Errorf("open uab file: %w", err)
		}
		section := bin.Section("linglong.meta")
		_, err = io.Copy(os.Stdout, section.Open())
		if err != nil {
			return fmt.Errorf("copy meta: %w", err)
		}
	default:
		return fmt.Errorf("file type %s is unsupported", args.InputFile)
	}
	return nil
}
