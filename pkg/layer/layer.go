package layer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/myml/linglong-tools/pkg/erofs"
	"github.com/myml/linglong-tools/pkg/tarutils"
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

type Layer struct {
	filepath    string
	meta        *types.LayerFileMetaInfo
	erofsOffset int64
	erofsSize   uint64
	signOffset  int64
}

func NewLayer(filepath string) (*Layer, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	var layerFile Layer
	layerFile.filepath = filepath
	// 读取信息
	layerFile.meta, err = ParseMetaInfo(f)
	if err != nil {
		return nil, fmt.Errorf("parse info: %w", err)
	}
	// 获取erofs偏移
	layerFile.erofsOffset, err = f.Seek(0, 1)
	if err != nil {
		return nil, fmt.Errorf("seek file: %w", err)
	}
	// 计算erofs大小
	if layerFile.meta.ErofsSize == 0 {
		finfo, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("stat file: %w", err)
		}
		layerFile.signOffset = 0
		layerFile.erofsSize = uint64(finfo.Size() - layerFile.erofsOffset)
	} else {
		layerFile.erofsSize = layerFile.meta.ErofsSize
		layerFile.signOffset = layerFile.erofsOffset + int64(layerFile.meta.ErofsSize)
	}

	return &layerFile, nil
}

// 为了和uab保持一致，以及后续可能的变动
func (l *Layer) Close() error {
	return nil
}

func (l *Layer) HasSign() bool {
	return l.signOffset > 0
}

func (l *Layer) Extract(outputDir string) error {
	fsck, err := erofs.NewErofsCmd()
	if err != nil {
		return fmt.Errorf("create fsck command failed: %w", err)
	}
	err = fsck.Extract(l.filepath, outputDir, l.erofsOffset, int64(l.erofsSize))
	if err != nil {
		return fmt.Errorf("extract erofs failed: %w", err)
	}
	if l.HasSign() {
		err = l.ExtractSign(filepath.Join(outputDir, "entries/share/deepin-elf-verify/.elfsign"))
		if err != nil {
			return fmt.Errorf("extract sign data failed: %w", err)
		}
	}
	return nil
}

func (l *Layer) ExtractSign(outputDir string) error {
	f, err := os.Open(l.filepath)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	_, err = f.Seek(l.signOffset, 0)
	if err != nil {
		return fmt.Errorf("seek sign data: %w", err)
	}
	err = tarutils.ExtractTar(f, outputDir)
	if err != nil {
		return fmt.Errorf("extract sign data: %w", err)
	}
	return nil
}

func (l *Layer) SaveErofs(outputFile string) error {
	f, err := os.Open(l.filepath)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	_, err = f.Seek(l.erofsOffset, 0)
	if err != nil {
		return fmt.Errorf("seek sign data: %w", err)
	}
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, f)
	if err != nil {
		return fmt.Errorf("copy erofs data to output file: %w", err)
	}
	return nil
}
