package cmd

import (
	"debug/elf"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/myml/linglong-tools/pkg/layer"
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
	_, err := exec.LookPath("erofsfuse")
	if err != nil {
		return errors.New("erofsfuse not found")
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

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("get status from output directory %s: %w", outputDir, err)
	}

	if len(entries) != 0 {
		return fmt.Errorf("output directory %s isn't empty", outputDir)
	}

	f, err := elf.Open(inputFile)
	if err != nil {
		return fmt.Errorf("open uab file error: %w", err)
	}
	defer f.Close()

	bundle := f.Section("linglong.bundle")
	if bundle == nil {
		return fmt.Errorf("%s doesn't has section linglong.bundle", inputFile)
	}

	mountPoint, err := os.MkdirTemp("", "uab-*")
	if err != nil {
		return fmt.Errorf("create temp mount point failed: %w", err)
	}
	defer os.RemoveAll(mountPoint)

	cmd := exec.Command("erofsfuse", fmt.Sprintf("--offset=%d", bundle.Offset), inputFile, mountPoint)
	if os.Getenv("LINGLONG_UAB_DEBUG") != "" {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("erofsfuse error: %w", err)
	}
	defer func() {
		cmd := exec.Command("umount", "-l", mountPoint)
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "please umount %s manually", mountPoint)
		}
	}()

	err = filepath.WalkDir(mountPoint, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error occurred while processing file %s: %w", path, err)
		}

		relative, err := filepath.Rel(mountPoint, path)
		if err != nil {
			return fmt.Errorf("error occurred while processing file %s: %w", path, err)
		}

		destination := filepath.Join(outputDir, relative)

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get original directory %s info: %w", path, err)
		}

		if d.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("failed to read symlink from %s: %w", path, err)
			}

			return os.Symlink(target, destination)
		}

		if d.IsDir() {
			err = os.MkdirAll(destination, info.Mode())
			if err != nil {
				return fmt.Errorf("failed to create destination directory %s: %w", destination, err)
			}

			return nil
		}

		dst, err := os.Create(destination)
		if err != nil {
			return fmt.Errorf("failed to create destination file %s: %w", destination, err)
		}
		defer dst.Close()

		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file %s: %w", path, err)
		}
		defer src.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", path, destination, err)
		}

		return os.Chmod(destination, info.Mode())
	})

	if err != nil {
		return err
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
