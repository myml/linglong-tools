package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/myml/linglong-tools/pkg/checker"
	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/myml/linglong-tools/pkg/types"
	"github.com/myml/linglong-tools/pkg/uab"
	"github.com/spf13/cobra"
)

type CheckArgs struct {
	InputFile           string
	FormatOutput        string
	CheckSigned         bool
	CheckSystemdService bool
	CheckIcon           bool
	CheckDesktopFile    bool
	CheckID             bool
}

type CheckResult struct {
	Pass           bool
	Signed         *CheckResultItem `json:",omitempty"`
	SystemdService *CheckResultItem `json:",omitempty"`
	Icon           *CheckResultItem `json:",omitempty"`
	ID             *CheckResultItem `json:",omitempty"`
	DesktopFile    *CheckResultItem `json:",omitempty"`
}

type CheckResultItem struct {
	Pass    bool
	Message string
}

func initCheckCmd() *cobra.Command {
	checkArgs := CheckArgs{}
	checkCmd := cobra.Command{
		Use:   "check",
		Short: "Package checker for linypas",
		Run: func(cmd *cobra.Command, args []string) {
			err := checkRun(checkArgs)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	checkCmd.Flags().StringVarP(&checkArgs.InputFile, "file", "f", "", "input file")
	checkCmd.Flags().BoolVar(&checkArgs.CheckSystemdService, "systemd", true, "check systemd service files")
	checkCmd.Flags().BoolVar(&checkArgs.CheckSigned, "signed", true, "check file is signed or not")
	checkCmd.Flags().BoolVar(&checkArgs.CheckIcon, "icon", true, "check icon")
	checkCmd.Flags().BoolVar(&checkArgs.CheckID, "id", true, "check app id is valid or not")
	checkCmd.Flags().BoolVar(&checkArgs.CheckDesktopFile, "desktop-file", true, "check desktop file is valid or not")
	checkCmd.Flags().StringVar(&checkArgs.FormatOutput, "format", "", "Format output using a custom template")
	if err := checkCmd.MarkFlagRequired("file"); err != nil {
		log.Fatalf("mark flag required: %v", err)
	}
	return &checkCmd
}

func checkRun(args CheckArgs) error {
	dir, err := os.MkdirTemp("", "ll-check-")
	if err != nil {
		return fmt.Errorf("create temp dir failed: %w", err)
	}
	defer os.RemoveAll(dir)

	extractor, err := newFileExtractor(args.InputFile)
	if err != nil {
		return err
	}
	defer extractor.Close()

	result := CheckResult{Pass: true}

	if args.CheckSigned {
		result.Signed = &CheckResultItem{Pass: extractor.HasSign()}
		if !result.Signed.Pass {
			result.Pass = false
		}
	}

	if args.CheckDesktopFile || args.CheckID || args.CheckIcon || args.CheckSystemdService {
		if err := extractor.Extract(dir); err != nil {
			return fmt.Errorf("extract file failed: %w", err)
		}

		root := os.DirFS(dir)
		info, err := loadLayerInfo(root)
		if err != nil {
			return err
		}

		id := info.ID
		if id == "" {
			id = info.Appid
		}

		if args.CheckID {
			if err := checker.CheckID(id); err != nil {
				result.ID = &CheckResultItem{Pass: false, Message: err.Error()}
				result.Pass = false
			} else {
				result.ID = &CheckResultItem{Pass: true}
			}
		}

		if args.CheckSystemdService {
			if err := checker.CheckSystemd(root); err != nil {
				result.SystemdService = &CheckResultItem{Pass: false, Message: err.Error()}
				result.Pass = false
			} else {
				result.SystemdService = &CheckResultItem{Pass: true}
			}
		}

		if args.CheckIcon {
			if err := checker.CheckIcon(root); err != nil {
				result.Icon = &CheckResultItem{Pass: false, Message: err.Error()}
				result.Pass = false
			} else {
				result.Icon = &CheckResultItem{Pass: true}
			}
		}

		if args.CheckDesktopFile {
			if err := checker.CheckDesktopFile(root, id); err != nil {
				result.DesktopFile = &CheckResultItem{Pass: false, Message: err.Error()}
				result.Pass = false
			} else {
				result.DesktopFile = &CheckResultItem{Pass: true}
			}
		}
	}

	return outputResult(result, args.FormatOutput)
}

type fileExtractor interface {
	HasSign() bool
	Extract(outputDir string) error
	Close() error
}

func newFileExtractor(filename string) (fileExtractor, error) {
	switch ext := filepath.Ext(filename); ext {
	case ".layer":
		layerFile, err := layer.NewLayer(filename)
		if err != nil {
			return nil, fmt.Errorf("create layer from file failed: %w", err)
		}
		return layerFile, nil
	case ".uab":
		uabFile, err := uab.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("open uab file: %w", err)
		}
		return uabFile, nil
	default:
		return nil, fmt.Errorf("file type %s is unsupported", filename)
	}
}

func loadLayerInfo(root fs.FS) (*types.LayerInfo, error) {
	data, err := fs.ReadFile(root, "files/info.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read from info.json: %w", err)
	}

	var info types.LayerInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal from info.json: %w", err)
	}

	return &info, nil
}

func outputResult(result CheckResult, format string) error {
	if format != "" {
		tmpl, err := template.New("").Parse(format)
		if err != nil {
			return fmt.Errorf("parse format: %w", err)
		}
		if err := tmpl.Execute(os.Stdout, result); err != nil {
			return fmt.Errorf("exec format: %w", err)
		}
		return nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	data = append(data, '\n')
	os.Stdout.Write(data)
	return nil
}
