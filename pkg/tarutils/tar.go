package tarutils

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	defer input.Close()
	_, err = io.Copy(tw, input)
	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}
	return nil
}

// 创建简单tar，用于打包签名
func CreateTar(w io.Writer, dir string) error {
	tw := tar.NewWriter(w)
	defer func() {
		if err := tw.Close(); err != nil {
			log.Fatal(err)
		}
	}()
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
		return fmt.Errorf("walk directory: %w", err)
	}
	return nil
}

// 解压简单tar，用于解压签名
func ExtractTar(r io.Reader, outputDir string) error {
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read sign data: %w", err)
		}
		out := filepath.Join(outputDir, header.Name)
		switch header.Typeflag {
		case tar.TypeReg:
			err = os.MkdirAll(filepath.Dir(out), 0755)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("ExtractTarGz: MkdirAll() failed: %w", err)
			}
			outFile, err := os.Create(out)
			if err != nil {
				return fmt.Errorf("extract sign data: Create() failed: %w", err)
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("extract sign data: Copy() failed: %w", err)
			}
		}
	}
	return nil
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
		w.Close()
	}()
	return r, nil
}
