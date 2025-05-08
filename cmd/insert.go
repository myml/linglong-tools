package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/myml/linglong-tools/pkg/tarutils"
	"github.com/myml/linglong-tools/pkg/uab"
	"github.com/spf13/cobra"
)

type InsertArgs struct {
	Update     bool
	InputFile  string
	SignDir    string
	OutputFile string
}

func initInsertCmd() *cobra.Command {
	var signArgs = InsertArgs{}
	signCmd := cobra.Command{
		Use:   "insert",
		Short: "Add sign files to linglong uab file",
		Example: `  # Insert sign files to uab file
  linglong-tools insert -f ./test.layer -d ./signs
  # Update sign data
  linglong-tools insert -f ./test.uab -d ./signs
  # Without changing the original file
  linglong-tools insert -f ./test.uab -d ./signs -o ./signed.uab
  `,
		Run: func(cmd *cobra.Command, args []string) {
			err := insertRun(signArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	signCmd.Flags().BoolVarP(&signArgs.Update, "update", "u", false, "update sign data")
	signCmd.Flags().StringVarP(&signArgs.InputFile, "file", "f", "", "uab file")
	signCmd.Flags().StringVarP(&signArgs.SignDir, "directory", "d", "", "sign directory")
	signCmd.Flags().StringVarP(&signArgs.OutputFile, "output", "o", "", "output file")
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

func insertRun(args InsertArgs) error {
	switch ext := filepath.Ext(args.InputFile); ext {
	case ".layer":
		return insert2Layer(args)
	case ".uab":
		return insert2UAB(args)
	default:
		return fmt.Errorf("file type %s is unsupported", args.InputFile)
	}
}

func preSignDiectory(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir %s failed: %w", dir, err)
	}
	for i := range files {
		name := files[i].Name()
		if len(name) <= 2 {
			continue
		}
		tierDir, tierName := filepath.Join(dir, name[:2]), name[2:]
		_, err = os.Stat(tierDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(tierDir, 0755)
				if err != nil {
					return fmt.Errorf("create tier dir %s failed: %w", tierDir, err)
				}
			} else {
				return fmt.Errorf("stat dir %s error: %w", tierDir, err)
			}
		}
		err = os.Rename(filepath.Join(dir, name), filepath.Join(tierDir, tierName))
		if err != nil {
			return fmt.Errorf("rename file %s to %s failed: %w", name, tierName, err)
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func insert2UAB(args InsertArgs) error {
	_, err := exec.LookPath("objcopy")
	if err != nil {
		return errors.New("objcopy not found")
	}
	info, err := os.Stat(args.SignDir)
	if err != nil {
		return fmt.Errorf("stat directory %s: %w", args.SignDir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s isn't a directory", args.SignDir)
	}

	uab, err := uab.Open(args.InputFile)
	if err != nil {
		return fmt.Errorf("open uab file: %w", err)
	}
	defer uab.Close()

	err = preSignDiectory(args.SignDir)
	if err != nil {
		return fmt.Errorf("preSignDiectory failed: %w", err)
	}
	if len(args.OutputFile) > 0 {
		err = CopyFile(args.InputFile, args.OutputFile)
		if err != nil {
			return fmt.Errorf("copy file failed: %w", err)
		}
	}
	// 如果不存在签名，则更新改为false，避免更新失败
	if args.Update && !uab.HasSign() {
		args.Update = false
	}
	return uab.InsertSign(args.SignDir, args.Update)
}

func insert2Layer(args InsertArgs) error {
	// 打开文件
	f, err := os.Open(args.InputFile)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	// 读取信息
	info, err := layer.ParseMetaInfo(f)
	if err != nil {
		return fmt.Errorf("parse info: %w", err)
	}
	err = preSignDiectory(args.SignDir)
	if err != nil {
		return fmt.Errorf("preSignDiectory failed: %w", err)
	}
	// 创建签名数据缓冲区
	var signDataBuff bytes.Buffer
	err = tarutils.CreateTar(&signDataBuff, args.SignDir)
	if err != nil {
		return fmt.Errorf("create tar file: %w", err)
	}
	info.SignSize = uint64(signDataBuff.Len())
	// 计算erofs大小
	if info.ErofsSize == 0 {
		offset, err := f.Seek(0, 1)
		if err != nil {
			return fmt.Errorf("seek file: %w", err)
		}
		finfo, err := f.Stat()
		if err != nil {
			return fmt.Errorf("stat file: %w", err)
		}
		info.ErofsSize = uint64(finfo.Size() - offset)
	}

	// 创建签名后的layer文件
	outdir := filepath.Dir(args.OutputFile)
	if len(args.OutputFile) == 0 {
		outdir = filepath.Dir(args.InputFile)
	}
	signed, err := os.CreateTemp(outdir, "signed_*.layer")
	if err != nil {
		return fmt.Errorf("open signed layer file: %w", err)
	}
	defer signed.Close()
	// 写入头信息
	n, err := signed.Write([]byte(info.Head))
	if err != nil {
		return fmt.Errorf("write head to signed layer file: %w", err)
	}
	_, err = signed.Write(bytes.Repeat([]byte{0}, 40-n))
	if err != nil {
		return fmt.Errorf("write padding to signed layer file: %w", err)
	}
	// 写入metainfo
	info.Raw = ""
	info.Head = ""
	meta, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal info: %w", err)
	}
	err = binary.Write(signed, binary.LittleEndian, uint32(len(meta)))
	if err != nil {
		return fmt.Errorf("binary write meta size: %w", err)
	}
	_, err = signed.Write(meta)
	if err != nil {
		return fmt.Errorf("write meta to signed layer file: %w", err)
	}
	// 写入erofs内容
	_, err = io.CopyN(signed, f, int64(info.ErofsSize))
	if err != nil {
		return fmt.Errorf("copy erofs content to signed layer file: %w", err)
	}
	// 写入签名数据
	_, err = signDataBuff.WriteTo(signed)
	if err != nil {
		return fmt.Errorf("write sign data to signed layer file: %w", err)
	}
	err = signed.Close()
	if err != nil {
		return fmt.Errorf("close signed layer file: %w", err)
	}
	err = os.Chmod(signed.Name(), 0644)
	if err != nil {
		return fmt.Errorf("change mode of signed layer file: %w", err)
	}
	if len(args.OutputFile) > 0 {
		err = os.Rename(signed.Name(), args.OutputFile)
	} else {
		err = os.Rename(signed.Name(), args.InputFile)
	}
	if err != nil {
		return fmt.Errorf("rename signed layer file: %w", err)
	}
	return nil
}
