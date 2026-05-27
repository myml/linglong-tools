package checker

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

var (
	validIconExtensions = []string{".png", ".svg", ".xpm"}
	validIconContexts   = []string{
		"actions", "apps", "categories", "devices", "emblems",
		"emotes", "filesystems", "intl", "mimetypes", "places",
		"status", "stock",
	}
)

func CheckIcon(root fs.FS) error {
	iconsDir := "files/share/icons"

	var iconFiles []string
	err := fs.WalkDir(root, iconsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !d.Type().IsRegular() {
			return fmt.Errorf("icon file %s is not a regular file", path)
		}

		ext := strings.ToLower(filepath.Ext(path))
		if slices.Contains(validIconExtensions, ext) {
			iconFiles = append(iconFiles, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk icons dir: %w", err)
	}

	if len(iconFiles) == 0 {
		return nil
	}

	for _, iconFile := range iconFiles {
		if err := validateIconPath(iconFile); err != nil {
			return fmt.Errorf("file %s: %w", filepath.Base(iconFile), err)
		}
	}

	return nil
}

func validateIconPath(path string) error {
	relPath := strings.TrimPrefix(path, "files/share/icons/")
	parts := strings.Split(relPath, string(filepath.Separator))

	if len(parts) < 3 {
		return errors.New("invalid icon path structure, expected theme/size/context/icon")
	}

	themeName := parts[0]
	if themeName == "" {
		return errors.New("theme name cannot be empty")
	}

	sizeDir := parts[1]
	if !isValidSizeDirectory(sizeDir) {
		return fmt.Errorf("invalid size directory %s, expected format like '48x48', 'scalable', or '48x48@2'", sizeDir)
	}

	context := parts[2]
	if !isValidContext(context) {
		return fmt.Errorf("invalid context %s", context)
	}

	return nil
}

func isValidSizeDirectory(dir string) bool {
	if dir == "scalable" {
		return true
	}

	if strings.HasPrefix(dir, "scalable@") {
		return true
	}

	if len(dir) >= 3 && dir[0] >= '0' && dir[0] <= '9' {
		xIndex := strings.Index(dir, "x")
		if xIndex > 0 {
			scaleIndex := strings.Index(dir, "@")
			if scaleIndex > xIndex {
				return true
			}
			if scaleIndex == -1 && xIndex < len(dir)-1 {
				return true
			}
		}
	}

	return false
}

func isValidContext(context string) bool {
	return slices.Contains(validIconContexts, context)
}
