package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

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
