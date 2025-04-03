package cmd

import (
	"errors"
	"fmt"
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
		layerFile, err := layer.NewLayer(args.File)
		if err != nil {
			return fmt.Errorf("create layer from file failed: %w", err)
		}
		defer layerFile.Close()
		return extract(args, layerFile)
	case ".uab":
		uabFile, err := uab.Open(args.File)
		if err != nil {
			return fmt.Errorf("open uab file: %w", err)
		}
		defer uabFile.Close()
		return extract(args, uabFile)
	default:
		return fmt.Errorf("file type %s is unsupported", args.File)
	}
}

type Extracter interface {
	Extract(outputDir string) error
	SaveErofs(outputFile string) error
}

func extract(args ExtractArgs, extracter Extracter) error {
	// 保存erofs镜像文件
	if len(args.OutputFile) > 0 {
		err := extracter.SaveErofs(args.OutputFile)
		if err != nil {
			return fmt.Errorf("save erofs file: %w", err)
		}
		return nil
	}
	if args.OutputDir == "" {
		return errors.New("please specific an output directory")
	}
	// 将erofs镜像解压到目录
	_, err := exec.LookPath("fsck.erofs")
	if err != nil {
		return errors.New("fsck.erofs not found")
	}
	_, err = os.Stat(args.OutputDir)
	if err == nil {
		return fmt.Errorf("output destination is already exist")
	}
	err = extracter.Extract(args.OutputDir)
	if err != nil {
		return fmt.Errorf("extract file to directory failed: %w", err)
	}
	return nil
}
