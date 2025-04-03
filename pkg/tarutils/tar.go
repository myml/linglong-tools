package tarutils

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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
		return appendFileToTar(dir, path, tw)
	})

	if err != nil {
		return fmt.Errorf("walk directory: %w", err)
	}
	return nil
}

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
