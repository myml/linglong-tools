package tarutils

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func appendFileToTar(root, file string, tw *tar.Writer) error {
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("create tar header failed: %w", err)
	}
	hdr.Name, err = filepath.Rel(root, file)
	if err != nil {
		return fmt.Errorf("rel path failed: %w", err)
	}
	hdr.Name = "./" + hdr.Name
	err = tw.WriteHeader(hdr)
	if err != nil {
		return fmt.Errorf("write header failed: %w", err)
	}
	if info.IsDir() {
		return nil
	}
	input, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}
	defer func() { _ = input.Close() }()
	_, err = io.Copy(tw, input)
	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}
	return nil
}

// 创建简单tar，用于打包签名
func CreateTar(w io.Writer, dir string) error {
	tw := tar.NewWriter(w)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if dir == path {
			return nil
		}
		return appendFileToTar(dir, path, tw)
	})
	if err != nil {
		_ = tw.Close()
		return fmt.Errorf("walk directory: %w", err)
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("close tar writer: %w", err)
	}
	return nil
}

// 解压简单tar，用于解压签名
func ExtractTar(r io.Reader, outputDir string) error {
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("read tar header: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		out, err := safeJoin(outputDir, header.Name)
		if err != nil {
			return fmt.Errorf("extract tar entry %q: %w", header.Name, err)
		}

		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return fmt.Errorf("create parent directory: %w", err)
		}

		if err := writeTarEntry(out, tarReader); err != nil {
			return err
		}
	}
	return nil
}

func writeTarEntry(path string, r io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %s: %w", path, err)
	}

	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return fmt.Errorf("write file %s: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close file %s: %w", path, err)
	}

	return nil
}

func safeJoin(base, name string) (string, error) {
	cleaned := filepath.Clean(name)
	if filepath.IsAbs(cleaned) || cleaned == ".." ||
		strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid path %q in tar", name)
	}

	return filepath.Join(base, cleaned), nil
}

// 使用tar命令行创建标准的tar.gz文件，支持硬链接、软链接等情况
func CreateTarStream(dir string) (io.ReadCloser, error) {
	r, w := io.Pipe()
	var stderrBuff bytes.Buffer
	cmd := exec.Command("tar", "-czvf", "-", "-C", dir, ".")
	cmd.Stdout = w
	cmd.Stderr = &stderrBuff
	go func() {
		err := cmd.Run()
		if err != nil {
			r.CloseWithError(fmt.Errorf("CreateTarCommand: tar command failed: %w\n%s", err, stderrBuff.String()))
		}
		_ = w.Close()
	}()
	return r, nil
}
