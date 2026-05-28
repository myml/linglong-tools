package checker

import (
	"testing"
)

func TestCheckID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid simple", "com.example.app", false},
		{"valid with numbers", "org.proj123.test", false},
		{"valid with underscore", "com.my_app.service", false},
		{"valid with hyphen in last", "com.example.my-app", false},
		{"valid three segments", "io.github.project", false},
		{"valid mixed case", "COM.Example.App", false},

		{"empty id", "", true},
		{"too long", string(make([]byte, 256)), true},
		{"single segment", "invalid", true},
		{"starts with digit", "1com.example", true},
		{"segment starts with digit", "com.1example.app", true},
		{"hyphen in non-last segment", "com-example.app", true},
		{"empty segment", "com..example", true},
		{"ends with dot", "com.example.", true},
		{"invalid char in last", "com.example.app$", true},
		{"last segment digit start", "com.example.1app", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func FuzzCheckID(f *testing.F) {
	seeds := []string{
		"com.example.app",
		"org.my-app_test",
		"io.github.project",
		"",
		"a.b",
		"1com.example",
		"com.example.1app",
		"com-example.app",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, id string) {
		err := CheckID(id)
		if err == nil {
			if id == "" {
				t.Error("expected error for empty id")
			}
			if len(id) > 255 {
				t.Error("expected error for id > 255 chars")
			}
		}
	})
}
