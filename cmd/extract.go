package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/spf13/cobra"
)

type ExtractArgs struct {
	LayerFile  string
	OutputFile string
}

// initExtractCmd用于提取layer文件的erofs镜像
func initExtractCmd() *cobra.Command {
	var extractArgs = ExtractArgs{}
	infoCmd := cobra.Command{
		Use:   "extract",
		Short: "Extract erofs img of linglong layer file",
		Example: `  # Use shell pipe redirect
  linglong-tools extract -f ./test.layer > app.img
  # Use output args
  linglong-tools extract -f ./test.layer -o app.img
  # Mount erofs img to /mnt/test
  erofsfuse app.img /mnt/test`,
		Run: func(cmd *cobra.Command, args []string) {
			err := extractRun(extractArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	infoCmd.Flags().StringVarP(&extractArgs.LayerFile, "file", "f", infoArgs.LayerFile, "layer file")
	infoCmd.Flags().StringVarP(&extractArgs.OutputFile, "output", "o", "/dev/stdout", "output file")
	err := infoCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	return &infoCmd
}

func extractRun(args ExtractArgs) error {
	// 打开文件
	f, err := os.Open(args.LayerFile)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	// 读取信息
	_, err = layer.ParseMetaInfo(f)
	if err != nil {
		return fmt.Errorf("parse info: %w", err)
	}
	var out io.Writer
	if args.OutputFile == "/dev/stdout" {
		out = os.Stdout
	} else {
		outFile, err := os.Create(args.OutputFile)
		if err != nil {
			return fmt.Errorf("create output: %w", err)
		}
		defer outFile.Close()
		out = outFile
	}
	_, err = io.Copy(out, f)
	if err != nil {
		return fmt.Errorf("copy content: %w", err)
	}
	return nil
}
