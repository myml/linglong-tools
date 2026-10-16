package uab

import (
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/myml/linglong-tools/v2/pkg/erofs"
	"github.com/myml/linglong-tools/v2/pkg/tarutils"
	"github.com/myml/linglong-tools/v2/pkg/types"
)

const (
	metaSectionName = "linglong.meta"
	bundleKey       = "bundle"
	signSectionName = "linglong.bundle.sign"
	signRelPath     = "entries/share/deepin-elf-verify/.elfsign"
)

type UAB struct {
	path     string
	metadata types.UABFileMetaInfo
}

func Open(file string) (*UAB, error) {
	if filepath.Ext(file) != ".uab" {
		return nil, fmt.Errorf("%s isn't a UAB file", file)
	}

	meta, err := readMetaInfo(file)
	if err != nil {
		return nil, err
	}

	return &UAB{path: file, metadata: meta}, nil
}

func readMetaInfo(path string) (types.UABFileMetaInfo, error) {
	var meta types.UABFileMetaInfo

	f, err := elf.Open(path)
	if err != nil {
		return meta, fmt.Errorf("open elf %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	section := f.Section(metaSectionName)
	if section == nil {
		return meta, fmt.Errorf("no %s section", metaSectionName)
	}

	data, err := io.ReadAll(section.Open())
	if err != nil {
		return meta, fmt.Errorf("read %s section: %w", metaSectionName, err)
	}

	if err := json.Unmarshal(data, &meta); err != nil {
		return meta, fmt.Errorf("unmarshal %s section: %w", metaSectionName, err)
	}

	return meta, nil
}

func (u *UAB) Close() error {
	return nil
}

func (u *UAB) MetaInfo() types.UABFileMetaInfo {
	return u.metadata
}

func (u *UAB) openELF() (*elf.File, error) {
	f, err := elf.Open(u.path)
	if err != nil {
		return nil, fmt.Errorf("open elf %s: %w", u.path, err)
	}

	return f, nil
}

func (u *UAB) bundleSection(f *elf.File) (*elf.Section, error) {
	name, exist := u.metadata.Sections[bundleKey]
	if !exist {
		return nil, fmt.Errorf("couldn't find %q section name in metadata", bundleKey)
	}

	section := f.Section(name)
	if section == nil {
		return nil, fmt.Errorf("couldn't find section %s in %s", name, u.path)
	}

	return section, nil
}

func (u *UAB) Extract(outputDir string) error {
	f, err := u.openELF()
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	bundle, err := u.bundleSection(f)
	if err != nil {
		return err
	}
	// 解压应用数据
	fsck, err := erofs.NewErofsCmd()
	if err != nil {
		return fmt.Errorf("create fsck command failed: %w", err)
	}

	if err := fsck.Extract(u.path, outputDir, int64(bundle.Offset), int64(bundle.Size)); err != nil {
		return fmt.Errorf("extract erofs failed: %w", err)
	}
	// 解压签名数据
	sign := f.Section(signSectionName)
	if sign == nil {
		return nil
	}

	appLayerPath, err := u.AppLayerPath()
	if err != nil {
		return fmt.Errorf("get app layer path error: %w", err)
	}

	signPath := filepath.Join(outputDir, appLayerPath, signRelPath)
	if err := os.MkdirAll(signPath, 0o755); err != nil {
		return fmt.Errorf("mkdir sign path: %w", err)
	}

	if err := tarutils.ExtractTar(sign.Open(), signPath); err != nil {
		return fmt.Errorf("extract uab sign: %w", err)
	}

	return nil
}

// InsertSign 插入签名到uab文件中
func (u *UAB) InsertSign(dataDir string, update bool) error {
	hasSign := u.HasSign()
	if hasSign && !update {
		return errors.New("sign section already exists, you could update it")
	}

	if !hasSign && update {
		return errors.New("no sign section to update")
	}

	signFile, err := os.CreateTemp("", "linglong-sign-*.tar")
	if err != nil {
		return fmt.Errorf("create sign file failed: %w", err)
	}

	defer func() { _ = os.Remove(signFile.Name()) }()
	defer func() { _ = signFile.Close() }()

	if err := tarutils.CreateTar(signFile, dataDir); err != nil {
		return fmt.Errorf("create tar file error: %w", err)
	}

	info, err := signFile.Stat()
	if err != nil {
		return fmt.Errorf("stat sign file error: %w", err)
	}

	if _, err := signFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek sign file error: %w", err)
	}

	target, err := os.OpenFile(u.path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open uab file for writing: %w", err)
	}
	defer func() { _ = target.Close() }()

	if err := writeELFSection(target, signSectionName, signFile, uint64(info.Size()), update); err != nil {
		return fmt.Errorf("insert sign section: %w", err)
	}

	return nil
}

// HasSign 返回uab是否包含签名
func (u *UAB) HasSign() bool {
	f, err := u.openELF()
	if err != nil {
		return false
	}

	defer func() { _ = f.Close() }()
	return f.Section(signSectionName) != nil
}

func (u *UAB) SignSize() uint64 {
	f, err := u.openELF()
	if err != nil {
		return 0
	}

	defer func() { _ = f.Close() }()
	if section := f.Section(signSectionName); section != nil {
		return section.Size
	}

	return 0
}

func (u *UAB) AppInfo() (types.LayerInfo, error) {
	for i := range u.metadata.Layers {
		if u.metadata.Layers[i].Info.Kind == "app" {
			return u.metadata.Layers[i].Info, nil
		}
	}

	return types.LayerInfo{}, errors.New("couldn't find app layer in layers")
}

func (u *UAB) AppLayerPath() (string, error) {
	info, err := u.AppInfo()
	if err != nil {
		return "", err
	}

	return filepath.Join("layers", info.ID, info.Module), nil
}

// ExtractSign 解压签名数据到指定目录
func (u *UAB) ExtractSign(outputDir string) error {
	f, err := u.openELF()
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	sign := f.Section(signSectionName)
	if sign == nil {
		return fmt.Errorf("couldn't find section %s in %s", signSectionName, u.path)
	}

	if err := tarutils.ExtractTar(sign.Open(), outputDir); err != nil {
		return fmt.Errorf("extract sign data: %w", err)
	}

	return nil
}

func (u *UAB) SaveErofs(outputFile string) error {
	f, err := u.openELF()
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	bundle, err := u.bundleSection(f)
	if err != nil {
		return err
	}

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, bundle.Open()); err != nil {
		return fmt.Errorf("write to output file: %w", err)
	}

	return nil
}
