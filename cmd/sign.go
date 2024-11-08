package cmd

import (
	"archive/tar"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

type SignArgs struct {
	File      string
	Directory string
}

func initSignCmd() *cobra.Command {
	var signArgs = SignArgs{}
	signCmd := cobra.Command{
		Use:   "sign",
		Short: "Add sign files to linglong uab file",
		Example: `  # Sign uab file
  linglong-tools sign -f ./test.uab -d ./signs`,
		Run: func(cmd *cobra.Command, args []string) {
			err := signRun(signArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	signCmd.Flags().StringVarP(&signArgs.File, "file", "f", "", "uab file")
	signCmd.Flags().StringVarP(&signArgs.Directory, "directory", "d", "", "sign directory")

	err := signCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}

	err = signCmd.MarkFlagRequired("directory")
	if err != nil {
		panic(err)
	}

	return &signCmd
}

func appendFileToTar(file string, tw *tar.Writer) error {
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}

	hdr := &tar.Header{
		Name: info.Name(),
		Mode: int64(info.Mode()),
		Size: info.Size(),
	}

	err = tw.WriteHeader(hdr)
	if err != nil {
		return fmt.Errorf("write header failed: %w", err)
	}

	input, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}
	defer input.Close()

	buf := make([]byte, 4096)
	for {
		bytes, err := input.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				return fmt.Errorf("read failed: %w", err)
			}
			break
		}

		_, err = tw.Write(buf[:bytes])
		if err != nil {
			return fmt.Errorf("write to tar failed: %w", err)
		}
	}

	return nil
}

func createTar(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read directory %s: %w", dir, err)
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("directory %s is empty", dir)
	}

	tarPath := filepath.Join(dir, "sign.tar")
	out, err := os.Create(tarPath)
	if err != nil {
		return "", fmt.Errorf("create tar file: %w", err)
	}
	defer out.Close()

	tw := tar.NewWriter(out)
	defer tw.Close()

	for _, entry := range entries {
		file := filepath.Join(dir, entry.Name())
		err = appendFileToTar(file, tw)
		if err != nil {
			return "", fmt.Errorf("while processing %s: %w", file, err)
		}
	}

	return tarPath, nil
}

func insertSignSection(uab string, tarPath string) error {
	_, err := exec.LookPath("objcopy")
	if err != nil {
		return errors.New("objcopy not found")
	}

	cmd := exec.Command("objcopy", "--add-section", fmt.Sprintf("linglong.bundle.sign=%s", tarPath),
		"--set-section-flags", "linglong.bundle.sign=readonly", uab, uab)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("run objcopy: %w", err)
	}

	return nil
}

func isElf(file string) (bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer f.Close()

	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil {
		return false, err
	}

	check := (magic[0] == 0x7F && magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F')
	return check, nil
}

func signRun(args SignArgs) error {
	file := args.File
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("stat file %s error: %w", file, err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("file %s isn't a regular file", file)
	}

	if filepath.Ext(file) != ".uab" {
		return fmt.Errorf("file type %s is unsupported", file)
	}

	b, err := isElf(file)
	if err != nil {
		return fmt.Errorf("check file error: %w", err)
	}

	if !b {
		return fmt.Errorf("input file type isn't elf: %s", file)
	}

	dir := args.Directory
	info, err = os.Stat(dir)
	if err != nil {
		return fmt.Errorf("stat directory %s: %w", dir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s isn't a directory", dir)
	}

	tarPath, err := createTar(dir)
	if err != nil {
		return fmt.Errorf("create tar file error: %w", err)
	}
	defer os.Remove(tarPath)

	err = insertSignSection(file, tarPath)
	if err != nil {
		return fmt.Errorf("insert sign section error: %w", err)
	}

	return nil
}
