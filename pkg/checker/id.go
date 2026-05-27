package checker

import (
	"errors"
	"fmt"
	"strings"
)

func isValidLastSegment(seg string) bool {
	for _, r := range seg {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
	}

	return true
}

func CheckID(id string) error {
	if id == "" {
		return errors.New("application ID cannot be empty")
	}

	if len(id) > 255 {
		return errors.New("application ID length cannot exceed 255 characters")
	}

	segments := strings.Split(id, ".")
	if len(segments) < 2 {
		return errors.New("application ID must contain at least two segments (e.g., com.example)")
	}

	for i := range len(segments) - 1 {
		seg := segments[i]
		if seg == "" {
			return errors.New("invalid format: empty segment or consecutive dots detected")
		}

		firstChar := seg[0]
		if (firstChar < 'a' || firstChar > 'z') && (firstChar < 'A' || firstChar > 'Z') {
			return fmt.Errorf("invalid segment '%s': each segment must start with a letter", seg)
		}

		for _, r := range seg {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
				return fmt.Errorf("invalid character in segment '%s': leading segments cannot contain hyphens '-'", seg)
			}
		}
	}

	last := segments[len(segments)-1]
	if last == "" {
		return errors.New("invalid format: application ID cannot end with a dot")
	}

	lastFirstChar := last[0]
	if (lastFirstChar < 'a' || lastFirstChar > 'z') && (lastFirstChar < 'A' || lastFirstChar > 'Z') {
		return fmt.Errorf("invalid last segment '%s': must start with a letter", last)
	}

	if !isValidLastSegment(last) {
		return fmt.Errorf("invalid character in last segment '%s': only alphanumeric, underscores, and hyphens are allowed", last)
	}

	return nil
}
