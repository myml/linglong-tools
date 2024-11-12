package cmd

import (
	"archive/tar"
	"debug/elf"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

type InsertArgs struct {
	Update    bool
	File      string
	Directory string
}

func initInsertCmd() *cobra.Command {
	var signArgs = InsertArgs{}
	signCmd := cobra.Command{
		Use:   "insert",
		Short: "Add sign files to linglong uab file",
		Example: `  # Insert sign files to uab file
  linglong-tools insert -f ./test.uab -d ./signs
  # Update sign data
  linglong-tools insert -f ./test.uab -d ./signs -u
  `,
		Run: func(cmd *cobra.Command, args []string) {
			err := signRun(signArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	signCmd.Flags().BoolVarP(&signArgs.Update, "update", "u", false, "update sign data")
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

func appendFileToTar(root string, file string, tw *tar.Writer) error {
	fmt.Printf("append file: %s", file)
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}

	group, err := user.Current()
	if err != nil {
		return fmt.Errorf("get current user failed: %w", err)
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return fmt.Errorf("get group id failed: %w", err)
	}

	uid, err := strconv.Atoi(group.Uid)
	if err != nil {
		return fmt.Errorf("get user id failed: %w", err)
	}

	relPath, err := filepath.Rel(root, file)
	if err != nil {
		return fmt.Errorf("get relative path failed: %w", err)
	}

	hdr := &tar.Header{
		Name:     relPath,
		Mode:     int64(info.Mode()),
		Size:     info.Size(),
		Gid:      gid,
		Uid:      uid,
		Uname:    group.Username,
		Gname:    group.Username,
		ModTime:  info.ModTime(),
		Format:   tar.FormatPAX,
		Typeflag: tar.TypeReg,
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
	tarPath := filepath.Join(dir, "sign.tar")
	out, err := os.Create(tarPath)
	if err != nil {
		return "", fmt.Errorf("create tar file: %w", err)
	}
	defer out.Close()

	tw := tar.NewWriter(out)
	defer func() {
		if err := tw.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && d.Name() != "sign.tar" {
			appendFileToTar(dir, path, tw)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("walk directory: %w", err)
	}

	return tarPath, nil
}

func insertInsertSection(uab string, tarPath string, update bool) error {
	_, err := exec.LookPath("objcopy")
	if err != nil {
		return errors.New("objcopy not found")
	}

	op := "--add-section"
	if update {
		op = "--update-section"
	}

	cmd := exec.Command("objcopy", op, fmt.Sprintf("linglong.bundle.sign=%s", tarPath),
		"--set-section-flags", "linglong.bundle.sign=readonly", uab, uab)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("run objcopy: %w", err)
	}

	return nil
}

func checkInsert(uab string) (bool, error) {
	f, err := elf.Open(uab)
	if err != nil {
		return false, err
	}
	defer f.Close()

	section := f.Section("linglong.bundle.sign")
	return section != nil, nil
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

func signRun(args InsertArgs) error {
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

	b, err = checkInsert(file)
	if err != nil {
		return fmt.Errorf("check sign error: %w", err)
	}

	if b && !args.Update {
		return fmt.Errorf("file %s has been signed, use -u to update", file)
	}

	tarPath, err := createTar(dir)
	if err != nil {
		return fmt.Errorf("create tar file error: %w", err)
	}
	defer os.Remove(tarPath)

	err = insertInsertSection(file, tarPath, args.Update)
	if err != nil {
		return fmt.Errorf("insert sign section error: %w", err)
	}

	return nil
}
