package checker

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestCheckIcon(t *testing.T) {
	makeMapFS := func(files map[string]string) fs.FS {
		m := make(fstest.MapFS, len(files))
		for path, content := range files {
			m[path] = &fstest.MapFile{Data: []byte(content)}
		}
		return m
	}

	tests := []struct {
		name    string
		root    fs.FS
		wantErr bool
	}{
		{
			name:    "no icons dir",
			root:    makeMapFS(nil),
			wantErr: true,
		},
		{
			name: "valid icon path",
			root: makeMapFS(map[string]string{
				"files/share/icons/hicolor/48x48/apps/app.png": "fake-icon",
			}),
			wantErr: false,
		},
		{
			name: "valid scalable icon",
			root: makeMapFS(map[string]string{
				"files/share/icons/hicolor/scalable/apps/app.svg": "fake-icon",
			}),
			wantErr: false,
		},
		{
			name: "invalid size directory",
			root: makeMapFS(map[string]string{
				"files/share/icons/hicolor/bad-size/apps/app.png": "fake-icon",
			}),
			wantErr: true,
		},
		{
			name: "invalid context",
			root: makeMapFS(map[string]string{
				"files/share/icons/hicolor/48x48/invalidctx/app.png": "fake-icon",
			}),
			wantErr: true,
		},
		{
			name: "too few path parts",
			root: makeMapFS(map[string]string{
				"files/share/icons/theme/icon.png": "fake-icon",
			}),
			wantErr: true,
		},
		{
			name: "non-icon extension ignored",
			root: makeMapFS(map[string]string{
				"files/share/icons/hicolor/48x48/apps/app.txt": "fake-icon",
			}),
			wantErr: false,
		},
		{
			name: "scalable@2x icon",
			root: makeMapFS(map[string]string{
				"files/share/icons/hicolor/scalable@2x/apps/app.svg": "fake-icon",
			}),
			wantErr: false,
		},
		{
			name: "non-regular file in icons dir",
			root: func() fs.FS {
				m := make(fstest.MapFS)
				m["files/share/icons/hicolor/48x48/apps/app.png"] = &fstest.MapFile{Data: []byte("icon"), Mode: fs.ModeIrregular}
				return m
			}(),
			wantErr: true,
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckIcon(tt.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckIcon() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIconPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid 48x48", "files/share/icons/hicolor/48x48/apps/app.png", false},
		{"valid scalable", "files/share/icons/hicolor/scalable/apps/app.svg", false},
		{"valid scalable@2x", "files/share/icons/hicolor/scalable@2x/apps/app.png", false},
		{"valid actions context", "files/share/icons/hicolor/48x48/actions/action.png", false},
		{"valid 16x16", "files/share/icons/hicolor/16x16/apps/app.png", false},

		{"too few parts", "files/share/icons/hicolor/48x48/app.png", true},
		{"empty theme", "files/share/icons//48x48/apps/app.png", true},
		{"invalid size", "files/share/icons/hicolor/bad/apps/app.png", true},
		{"invalid context", "files/share/icons/hicolor/48x48/invalid/app.png", true},
		{"size without x", "files/share/icons/hicolor/480/apps/app.png", true},
		{"invalid char in size", "files/share/icons/hicolor/abc/apps/app.png", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIconPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIconPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidSizeDirectory(t *testing.T) {
	tests := []struct {
		dir   string
		valid bool
	}{
		{"scalable", true},
		{"scalable@2x", true},
		{"scalable@3x", true},
		{"48x48", true},
		{"16x16", true},
		{"256x256", true},
		{"48x48@2", true},
		{"22x22@4", true},

		{"", false},
		{"bad", false},
		{"48", false},
		{"x48", false},
		{"48x", false},
		{"48x48x48", true},
		{"abcxdef", false},
	}
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			got := isValidSizeDirectory(tt.dir)
			if got != tt.valid {
				t.Errorf("isValidSizeDirectory(%q) = %v, want %v", tt.dir, got, tt.valid)
			}
		})
	}
}

func TestIsValidContext(t *testing.T) {
	tests := []struct {
		context string
		valid   bool
	}{
		{"apps", true},
		{"actions", true},
		{"categories", true},
		{"devices", true},
		{"emblems", true},
		{"emotes", true},
		{"filesystems", true},
		{"intl", true},
		{"mimetypes", true},
		{"places", true},
		{"status", true},
		{"stock", true},
		{"invalid", false},
		{"APPS", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.context, func(t *testing.T) {
			got := isValidContext(tt.context)
			if got != tt.valid {
				t.Errorf("isValidContext(%q) = %v, want %v", tt.context, got, tt.valid)
			}
		})
	}
}

func FuzzIsValidSizeDirectory(f *testing.F) {
	seeds := []string{
		"scalable",
		"scalable@2x",
		"48x48",
		"16x16",
		"256x256@2",
		"bad",
		"",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, dir string) {
		isValidSizeDirectory(dir)
	})
}
