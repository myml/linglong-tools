package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/myml/linglong-tools/pkg/uab"
	"github.com/spf13/cobra"
)

type ExtractArgs struct {
	File       string
	OutputFile string
	OutputDir  string
}

// initExtractCmd用于提取layer文件的erofs镜像
func initExtractCmd() *cobra.Command {
	var extractArgs = ExtractArgs{}
	extractCmd := cobra.Command{
		Use:   "extract",
		Short: "Extract erofs img of linglong layer file or bundle of linglong uab file",
		Example: `  # Extract erofs to an empty directory
  linglong-tools extract -f ./test.uab -d /path/to/dir
  linglong-tools extract -f ./test.layer -d /path/to/dir
  # output erofs image
  linglong-tools extract -f ./test.uab -o app.img
  linglong-tools extract -f ./test.layer -o app.img
  # Mount erofs img to /tmp/test
  erofsfuse app.img /tmp/test`,
		Run: func(cmd *cobra.Command, args []string) {
			err := extractRun(extractArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	extractCmd.Flags().StringVarP(&extractArgs.File, "file", "f", "", "layer file")
	extractCmd.Flags().StringVarP(&extractArgs.OutputFile, "output", "o", "", "output file")
	extractCmd.Flags().StringVarP(&extractArgs.OutputDir, "dir", "d", "", "output directory")
	err := extractCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}

	return &extractCmd
}

func extractRun(args ExtractArgs) error {
	switch ext := filepath.Ext(args.File); ext {
	case ".layer":
		return extractLayer(args)
	case ".uab":
		return extractUab(args)
	default:
		return fmt.Errorf("file type %s is unsupported", args.File)
	}
}

func extractUab(args ExtractArgs) error {
	uabFile, err := uab.Open(args.File)
	if err != nil {
		return fmt.Errorf("open uab file: %w", err)
	}
	defer uabFile.Close()
	// 保存erofs镜像文件
	if len(args.OutputFile) > 0 {
		err = uabFile.SaveErofs(args.OutputFile)
		if err != nil {
			return fmt.Errorf("save uab file to erofs: %w", err)
		}
		return nil
	}
	// 将erofs镜像解压到目录
	outputDir := args.OutputDir
	_, err = exec.LookPath("fsck.erofs")
	if err != nil {
		return errors.New("fsck.erofs not found")
	}

	if outputDir == "" {
		return errors.New("please specific an output directory")
	}

	info, err := os.Stat(outputDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("create directory %s error: %w", outputDir, err)
		}
	} else if !info.IsDir() {
		return errors.New("output destination isn't a directory")
	}

	entries, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("get status from output directory %s: %w", outputDir, err)
	}

	if len(entries) != 0 {
		return fmt.Errorf("output directory %s isn't empty", outputDir)
	}

	err = uabFile.Extract(outputDir)
	if err != nil {
		return fmt.Errorf("extract uab file: %w", err)
	}
	return nil
}

func extractLayer(args ExtractArgs) error {
	f, err := os.Open(args.File)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	// 读取信息
	_, err = layer.ParseMetaInfo(f)
	if err != nil {
		return fmt.Errorf("parse info: %w", err)
	}
	// 保存erofs镜像文件
	if len(args.OutputFile) > 0 {
		outFile, err := os.Create(args.OutputFile)
		if err != nil {
			return fmt.Errorf("create output: %w", err)
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, f)
		if err != nil {
			return fmt.Errorf("copy content: %w", err)
		}
		return nil
	}
	offset, err := f.Seek(0, 1)
	if err != nil {
		return fmt.Errorf("seek file: %w", err)
	}
	log.Println("offset", offset)
	err = layer.Extract(args.File, offset, args.OutputDir)
	if err != nil {
		return fmt.Errorf("extract erofs: %w", err)
	}
	return nil
}
