package erofs

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type ErofsFsck struct {
	SupportOffset bool
}

func NewErofsCmd() (*ErofsFsck, error) {
	// 检查所需命令行是否存在
	_, err := exec.LookPath("fsck.erofs")
	if err != nil {
		return nil, errors.New("fsck.erofs not found")
	}
	out, err := exec.Command("fsck.erofs", "--help").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exec fsck.erofs --help error: %w %s", err, out)
	}
	cmd := ErofsFsck{
		SupportOffset: strings.Contains(string(out), "--offset="),
	}
	return &cmd, nil
}

// Extract 使用fsck.erofs解压erofs镜像文件
func (fsck *ErofsFsck) Extract(imgFile, outDir string, offset, length int64) error {
	if fsck.SupportOffset {
		out, err := exec.Command("fsck.erofs",
			fmt.Sprintf("--offset=%d", offset),
			fmt.Sprintf("--extract=%s", outDir),
			imgFile,
		).CombinedOutput()
		if err != nil {
			return fmt.Errorf("fsck erofs failed: %w %s", err, out)
		}
		return nil
	}
	// 不支持offset功能的fsck.erofs，将镜像文件另存再解压
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		return fmt.Errorf("create temp file error: %w", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	rawFile, err := os.Open(imgFile)
	if err != nil {
		return fmt.Errorf("open raw file error: %w", err)
	}
	defer rawFile.Close()
	_, err = rawFile.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("seek offset error: %w", err)
	}
	_, err = io.CopyN(tmpFile, rawFile, length)
	if err != nil {
		return fmt.Errorf("copy data error: %w", err)
	}
	err = tmpFile.Close()
	if err != nil {
		return fmt.Errorf("close temp file error: %w", err)
	}
	out, err := exec.Command("fsck.erofs",
		fmt.Sprintf("--extract=%s", outDir),
		tmpFile.Name(),
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("fsck erofs failed: %w %s", err, out)
	}
	return nil
}
