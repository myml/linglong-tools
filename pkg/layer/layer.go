package layer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/myml/linglong-tools/pkg/types"
)

// ParseMetaInfo parse layer metainfo
// layer file format <head(chars 40 byte)> <json payload size(uint32 4 byte)> <json payload(varchar)> <erofs image> <sign data>
func ParseMetaInfo(r io.Reader) (*types.LayerFileMetaInfo, error) {
	var buff bytes.Buffer
	_, err := io.CopyN(&buff, r, int64(40))
	if err != nil {
		return nil, fmt.Errorf("read head: %w", err)
	}
	head := buff.String()
	buff.Reset()
	var size uint32
	err = binary.Read(r, binary.LittleEndian, &size)
	if err != nil {
		return nil, fmt.Errorf("read info size: %w", err)
	}
	_, err = io.CopyN(&buff, r, int64(size))
	if err != nil {
		return nil, fmt.Errorf("read info data: %w", err)
	}
	var info types.LayerFileMetaInfo
	err = json.Unmarshal(buff.Bytes(), &info)
	if err != nil {
		return nil, fmt.Errorf("unmarshal info data: %w", err)
	}
	if len(info.Info.ID) > 0 {
		info.Info.Appid = info.Info.ID
	} else {
		info.Info.ID = info.Info.Appid
	}
	if len(info.Info.Appid) == 0 {
		return nil, fmt.Errorf("missing appid field in raw(%s)", buff.String())
	}
	info.Head = strings.TrimSpace(head)
	info.Raw = buff.String()
	return &info, nil
}

func Extract(filepath string, offset int64, outputDir string) error {
	cmd := exec.Command("fsck.erofs",
		fmt.Sprintf("--offset=%d", offset),
		fmt.Sprintf("--extract=%s", outputDir),
		filepath,
	)
	if os.Getenv("LINGLONG_UAB_DEBUG") != "" {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("fsck erofs failed: %w", err)
		}
		return nil
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fsck erofs failed: %w, %s", err, out)
	}
	return nil
}
