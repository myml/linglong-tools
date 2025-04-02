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
		Example: `  # Extract uab file to an empty directory
  linglong-tools extract -f ./test.uab -d /path/to/dir
  # Use shell pipe redirect
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

	extractCmd.Flags().StringVarP(&extractArgs.File, "file", "f", "", "layer file")
	extractCmd.Flags().StringVarP(&extractArgs.OutputFile, "output", "o", "/dev/stdout", "output file")
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
		return extractLayer(args.File, args.OutputFile)
	case ".uab":
		return extractUab(args.File, args.OutputDir)
	default:
		return fmt.Errorf("file type %s is unsupported", args.File)
	}
}

func extractUab(inputFile string, outputDir string) error {
	_, err := exec.LookPath("fsck.erofs")
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

	uabFile, err := uab.Open(inputFile)
	if err != nil {
		return fmt.Errorf("open uab file: %w", err)
	}
	defer uabFile.Close()

	err = uabFile.Extract(outputDir)
	if err != nil {
		return fmt.Errorf("extract uab file: %w", err)
	}
	return nil
}

func extractLayer(inputFile string, outputFile string) error {
	f, err := os.Open(inputFile)
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
	if outputFile == "/dev/stdout" {
		out = os.Stdout
	} else {
		outFile, err := os.Create(outputFile)
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
