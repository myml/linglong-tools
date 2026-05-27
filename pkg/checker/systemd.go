package checker

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var systemdUnitTypes = []string{".service", ".socket", ".timer", ".path", ".slice", ".target"}

func CheckSystemd(root fs.FS) error {
	systemdDir := "files/lib/systemd"

	var files []string
	err := fs.WalkDir(root, systemdDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !d.Type().IsRegular() {
			return fmt.Errorf("%s is not a regular file", path)
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("walk systemd dir: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		filename := filepath.Base(file)
		ext := strings.ToLower(filepath.Ext(filename))
		if !isValidSystemdUnitType(ext) {
			return fmt.Errorf("file %s has unsupported systemd unit type %s", filename, ext)
		}

		if err := validateServiceFilename(filename); err != nil {
			return fmt.Errorf("file %s: %w", filename, err)
		}

		if ext == ".service" || ext == ".socket" {
			if err := validateServiceContent(root, file); err != nil {
				return fmt.Errorf("file %s: %w", filename, err)
			}
		}
	}

	return nil
}

func isValidSystemdUnitType(ext string) bool {
	return slices.Contains(systemdUnitTypes, ext)
}

func isAlphaOrUnderscore(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isValidSystemdChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
		r == '_' || r == '-' || r == '@' || r == '.'
}

func validateServiceFilename(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	name := strings.TrimSuffix(filename, ext)
	if name == "" {
		return errors.New("unit filename cannot be empty")
	}

	for i, r := range name {
		if i == 0 {
			if !isAlphaOrUnderscore(r) {
				return errors.New("filename must start with letter or underscore")
			}
		} else {
			if !isValidSystemdChar(r) {
				return fmt.Errorf("filename contains invalid character '%c'", r)
			}
		}
	}

	return nil
}

func validateServiceContent(root fs.FS, file string) error {
	f, err := root.Open(file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Exec") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			value := strings.TrimSpace(parts[1])
			if value == "" {
				continue
			}

			if strings.HasPrefix(value, "/") {
				return fmt.Errorf("line %d: %s uses absolute path %s", lineNum, parts[0], value)
			}

			if parts[0] == "ExecStart" {
				execPath := parseExecPath(value)
				if execPath != "" && !isBinaryExists(root, execPath) {
					return fmt.Errorf("line %d: ExecStart binary %s not found in files/bin", lineNum, execPath)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	return nil
}

func parseExecPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	for i, r := range value {
		if r == ' ' || r == '\t' {
			return value[:i]
		}
	}
	return value
}

func isBinaryExists(root fs.FS, binary string) bool {
	path := "files/bin/" + binary
	_, err := fs.Stat(root, path)
	return err == nil
}
