package checker

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CheckDesktopFile(root fs.FS, id string) error {
	desktopDir := "files/share/applications"

	var files []string
	err := fs.WalkDir(root, desktopDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !d.Type().IsRegular() {
			return fmt.Errorf("%s is not a regular file", path)
		}

		if strings.HasSuffix(path, ".desktop") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("walk desktop dir: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		if err := validateDesktopFilename(file); err != nil {
			return fmt.Errorf("file %s: %w", filepath.Base(file), err)
		}

		if err := validateDesktopContent(root, file); err != nil {
			return fmt.Errorf("file %s: %w", filepath.Base(file), err)
		}
	}

	hasMatch := false
	for _, file := range files {
		filename := filepath.Base(file)
		name := strings.TrimSuffix(filename, ".desktop")

		if name == id {
			hasMatch = true
			break
		}
	}

	if !hasMatch {
		log.Printf("WARNING: no desktop file matches app id %s", id)
	}

	return nil
}

func validateDesktopContent(root fs.FS, file string) error {
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
		if !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if value == "" {
			continue
		}

		if key != "Exec" && key != "TryExec" && key != "Icon" {
			continue
		}

		if strings.HasPrefix(value, "/") {
			return fmt.Errorf("line %d: key %s has absolute path value %s", lineNum, key, value)
		}
	}
	return scanner.Err()
}

func validateDesktopFilename(file string) error {
	filename := filepath.Base(file)
	name := strings.TrimSuffix(filename, ".desktop")
	if name == "" {
		return fmt.Errorf("desktop file name cannot be empty")
	}

	// Must be a valid D-Bus well-known name:
	// sequence of non-empty elements separated by dots,
	// none starting with a digit, each containing only [A-Za-z0-9-_]
	segments := strings.Split(name, ".")
	for i, seg := range segments {
		if seg == "" {
			return fmt.Errorf("segment %d is empty in name %s", i, name)
		}

		if seg[0] >= '0' && seg[0] <= '9' {
			return fmt.Errorf("segment '%s' in name %s starts with a digit", seg, name)
		}

		for _, r := range seg {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' {
				return fmt.Errorf("invalid character '%c' in segment '%s' of name %s", r, seg, name)
			}
		}
	}

	return nil
}
