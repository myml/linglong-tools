package uab

import (
	"archive/tar"
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/myml/linglong-tools/pkg/tarutils"
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

// Extract 解压应用数据到指定目录，同时会解压签名文件
func (u *UAB) Extract(outputDir string) error {
	meta := u.MetaInfo()
	bundleSectionName, exist := meta.Sections["bundle"]
	if !exist {
		return errors.New("couldn't find bundle section name")
	}
	// 解压应用数据
	bundle := u.elf.Section(bundleSectionName)
	if bundle == nil {
		return fmt.Errorf("couldn't find section %s in %s", bundleSectionName, u.path)
	}

	cmd := exec.Command("fsck.erofs",
		fmt.Sprintf("--offset=%d", bundle.Offset),
		fmt.Sprintf("--extract=%s", outputDir),
		u.path,
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

	// 解压签名数据
	if u.HasSign() {
		appLayerPath, err := u.AppLayerPath()
		if err != nil {
			return fmt.Errorf("get app layer path error: %w", err)
		}
		appLayerPath = filepath.Join(outputDir, appLayerPath)
		signPath := filepath.Join(appLayerPath, "entries/share/deepin-elf-verify/.elfsign")
		err = os.MkdirAll(signPath, 0755)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("mkdir sign path: %w", err)
		}
		err = u.ExtractSign(signPath)
		if err != nil {
			return fmt.Errorf("extract uab sign: %w", err)
		}
	}

	return nil
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

// InsertSign 插入签名到uab文件中
func (u *UAB) InsertSign(dataDir string, update bool) error {
	signSection := u.elf.Section("linglong.bundle.sign")
	if signSection != nil && !update {
		return errors.New("sign section already exists, you could update it")
	}

	signFile, err := ioutil.TempFile("", "0600")
	if err != nil {
		return fmt.Errorf("create sign file failed: %w", err)
	}
	defer func() {
		signFile.Close()
		os.Remove(signFile.Name())
	}()
	err = tarutils.CreateTar(signFile, dataDir)
	if err != nil {
		return fmt.Errorf("create tar file error: %w", err)
	}
	err = signFile.Close()
	if err != nil {
		return fmt.Errorf("close sign file error: %w", err)
	}
	return u.insertSection(signFile.Name(), update)
}

// HasSign 返回uab是否包含签名
func (u *UAB) HasSign() bool {
	bundleSectionName := "linglong.bundle.sign"
	bundleSign := u.elf.Section(bundleSectionName)
	return bundleSign != nil
}

// AppLayerPath 返回应用层的路径
func (u *UAB) AppInfo() (types.LayerInfo, error) {
	meta := u.MetaInfo()
	layers := meta.Layers
	for i := range layers {
		if layers[i].Info.Kind == "app" {
			return layers[i].Info, nil
		}
	}
	return types.LayerInfo{}, fmt.Errorf("couldn't find app layer in layers")
}

// AppLayerPath 返回应用层的路径
func (u *UAB) AppLayerPath() (string, error) {
	info, err := u.AppInfo()
	if err != nil {
		return "", err
	}
	return filepath.Join("layers", info.ID, info.Module), nil
}

// ExtractSign 解压签名数据到指定目录
func (u *UAB) ExtractSign(outputDir string) error {
	bundleSectionName := "linglong.bundle.sign"
	bundleSign := u.elf.Section(bundleSectionName)
	if bundleSign == nil {
		return fmt.Errorf("couldn't find section %s in %s", bundleSectionName, u.path)
	}
	r := bundleSign.Open()
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

func (u *UAB) SaveErofs(outputFile string) error {
	meta := u.MetaInfo()
	bundleSectionName, exist := meta.Sections["bundle"]
	if !exist {
		return errors.New("couldn't find bundle section name")
	}
	// 解压应用数据
	bundle := u.elf.Section(bundleSectionName)
	if bundle == nil {
		return fmt.Errorf("couldn't find section %s in %s", bundleSectionName, u.path)
	}
	f, err := os.Open(outputFile)
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	defer f.Close()
	r := bundle.Open()
	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("write to output file: %w", err)
	}
	return nil
}
