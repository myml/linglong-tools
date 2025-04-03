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
	Update    bool
	File      string
	Directory string
}

func initInsertCmd() *cobra.Command {
	var signArgs = InsertArgs{}
	signCmd := cobra.Command{
		Use:   "insert",
		Short: "Add sign files to linglong uab file",
		Example: `  # Insert sign files to uab file
  linglong-tools insert -f ./test.uab -d ./signs
  # Update sign data
  linglong-tools insert -f ./test.uab -d ./signs -u
  `,
		Run: func(cmd *cobra.Command, args []string) {
			err := insertRun(signArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	signCmd.Flags().BoolVarP(&signArgs.Update, "update", "u", false, "update sign data")
	signCmd.Flags().StringVarP(&signArgs.File, "file", "f", "", "uab file")
	signCmd.Flags().StringVarP(&signArgs.Directory, "directory", "d", "", "sign directory")

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
	switch ext := filepath.Ext(args.File); ext {
	case ".layer":
		return insert2Layer(args)
	case ".uab":
		return insert2UAB(args)
	default:
		return fmt.Errorf("file type %s is unsupported", args.File)
	}
}

func insert2UAB(args InsertArgs) error {
	_, err := exec.LookPath("objcopy")
	if err != nil {
		return errors.New("objcopy not found")
	}

	info, err := os.Stat(args.Directory)
	if err != nil {
		return fmt.Errorf("stat directory %s: %w", args.Directory, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s isn't a directory", args.Directory)
	}

	uab, err := uab.Open(args.File)
	if err != nil {
		return fmt.Errorf("open uab file: %w", err)
	}
	defer uab.Close()

	return uab.InsertSign(args.Directory, args.Update)
}

func insert2Layer(args InsertArgs) error {
	// 打开文件
	f, err := os.Open(args.File)
	if err != nil {
		return fmt.Errorf("open layer file: %w", err)
	}
	defer f.Close()
	// 读取信息
	info, err := layer.ParseMetaInfo(f)
	if err != nil {
		return fmt.Errorf("parse info: %w", err)
	}
	// 创建签名数据缓冲区
	var signDataBuff bytes.Buffer
	err = tarutils.CreateTar(&signDataBuff, args.Directory)
	if err != nil {
		return fmt.Errorf("create tar file: %w", err)
	}
	info.SignSize = int64(signDataBuff.Len())
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
		info.ErofsSize = finfo.Size() - offset
	}
	// 创建签名后的layer文件
	signed, err := os.Create("signed.layer")
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
	_, err = io.CopyN(signed, f, info.ErofsSize)
	if err != nil {
		return fmt.Errorf("copy erofs content to signed layer file: %w", err)
	}
	// 写入签名数据
	_, err = signDataBuff.WriteTo(signed)
	if err != nil {
		return fmt.Errorf("write sign data to signed layer file: %w", err)
	}
	return nil
}
