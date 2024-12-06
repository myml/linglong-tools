package uab

import (
	"archive/tar"
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/myml/linglong-tools/pkg/types"
)

type UAB struct {
	path     string
	file     *os.File
	elf      *elf.File
	metadata types.UABFileMetaInfo
}

func Open(file string) (*UAB, error) {
	absPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}

	if filepath.Ext(file) != ".uab" {
		return nil, fmt.Errorf("%s isn't a UAB file", absPath)
	}

	elf, err := elf.Open(absPath)
	if err != nil {
		return nil, err
	}

	metaInfo := elf.Section("linglong.meta")
	if metaInfo == nil {
		return nil, errors.New("no linglong.meta section")
	}

	s := io.NewSectionReader(f, int64(metaInfo.Offset), int64(metaInfo.Size))
	buf := make([]byte, metaInfo.Size)
	bytesRead, err := s.Read(buf)
	if err != nil {
		return nil, err
	}

	if bytesRead != int(metaInfo.Size) {
		return nil, errors.New("read linglong.meta section failed")
	}

	var meta types.UABFileMetaInfo
	err = json.Unmarshal(buf, &meta)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal %s: %w", buf, err)
	}

	return &UAB{absPath, f, elf, meta}, nil
}

func (u *UAB) Close() error {
	err := u.file.Close()
	if err != nil {
		return err
	}

	return u.elf.Close()
}

func (u *UAB) MetaInfo() types.UABFileMetaInfo {
	return u.metadata
}

func (u *UAB) Extract(outputDir string) error {
	meta := u.MetaInfo()
	bundleSectionName, exist := meta.Sections["bundle"]
	if !exist {
		return errors.New("couldn't find bundle section name")
	}

	bundle := u.elf.Section(bundleSectionName)
	if bundle == nil {
		return fmt.Errorf("couldn't find section %s in %s", bundleSectionName, u.path)
	}

	mountPoint, err := ioutil.TempDir("", "uab-*")
	if err != nil {
		return fmt.Errorf("create temp mount point failed: %w", err)
	}
	defer os.RemoveAll(mountPoint)

	cmd := exec.Command("erofsfuse", fmt.Sprintf("--offset=%d", bundle.Offset), u.path, mountPoint)
	if os.Getenv("LINGLONG_UAB_DEBUG") != "" {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("erofsfuse error: %w", err)
	}
	defer func() {
		cmd := exec.Command("fusermount", "-u", mountPoint)
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "please use fusermount to umount %s manually", mountPoint)
		}
	}()

	err = filepath.Walk(mountPoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error occurred while processing file %s: %w", path, err)
		}

		relative, err := filepath.Rel(mountPoint, path)
		if err != nil {
			return fmt.Errorf("error occurred while processing file %s: %w", path, err)
		}

		destination := filepath.Join(outputDir, relative)
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("failed to read symlink from %s: %w", path, err)
			}

			return os.Symlink(target, destination)
		}

		if info.IsDir() {
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

func appendFileToTar(root string, file string, tw *tar.Writer) error {
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

	parent := filepath.Dir(relPath)
	name := info.Name()
	targetPath := filepath.Join(parent, name[0:2], name)

	hdr := &tar.Header{
		Name:     targetPath,
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

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() != "sign.tar" {
			return appendFileToTar(dir, path, tw)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("walk directory: %w", err)
	}

	return tarPath, nil
}

func (u *UAB) insertSection(tarPath string, update bool) error {
	op := "--add-section"
	if update {
		op = "--update-section"
	}

	cmd := exec.Command("objcopy", op, fmt.Sprintf("linglong.bundle.sign=%s", tarPath),
		"--set-section-flags", "linglong.bundle.sign=readonly", u.path, u.path)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("run objcopy: %w", err)
	}

	return nil
}

func (u *UAB) InsertSign(dataDir string, update bool) error {
	signSection := u.elf.Section("linglong.bundle.sign")
	if signSection != nil && !update {
		return errors.New("sign section already exists, you could update it")
	}

	tarPath, err := createTar(dataDir)
	if err != nil {
		return fmt.Errorf("create tar file error: %w", err)
	}
	defer os.Remove(tarPath)

	return u.insertSection(tarPath, update)
}
