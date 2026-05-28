package checker

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestCheckDesktopFile(t *testing.T) {
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
		id      string
		wantErr bool
	}{
		{
			name:    "no desktop dir",
			root:    makeMapFS(nil),
			id:      "com.example.app",
			wantErr: false,
		},
		{
			name: "valid desktop file matching id",
			root: makeMapFS(map[string]string{
				"files/share/applications/com.example.app.desktop": "[Desktop Entry]\nExec=app\nIcon=app\n",
			}),
			id:      "com.example.app",
			wantErr: false,
		},
		{
			name: "desktop file with absolute Exec path",
			root: makeMapFS(map[string]string{
				"files/share/applications/com.example.app.desktop": "[Desktop Entry]\nExec=/usr/bin/app\n",
			}),
			id:      "com.example.app",
			wantErr: true,
		},
		{
			name: "desktop file with absolute Icon path",
			root: makeMapFS(map[string]string{
				"files/share/applications/com.example.app.desktop": "[Desktop Entry]\nIcon=/usr/share/icons/app.png\n",
			}),
			id:      "com.example.app",
			wantErr: true,
		},
		{
			name: "invalid desktop filename empty",
			root: makeMapFS(map[string]string{
				"files/share/applications/.desktop": "[Desktop Entry]\n",
			}),
			id:      "com.example.app",
			wantErr: true,
		},
		{
			name: "invalid desktop filename starts with digit",
			root: makeMapFS(map[string]string{
				"files/share/applications/1com.example.app.desktop": "[Desktop Entry]\n",
			}),
			id:      "com.example.app",
			wantErr: true,
		},
		{
			name: "invalid char in desktop filename",
			root: makeMapFS(map[string]string{
				"files/share/applications/com.example.app$.desktop": "[Desktop Entry]\n",
			}),
			id:      "com.example.app",
			wantErr: true,
		},
		{
			name: "empty desktop dir ignored",
			root: makeMapFS(map[string]string{
				"files/share/applications/.empty": "",
			}),
			id:      "com.example.app",
			wantErr: false,
		},
		{
			name: "multiple desktop files one invalid",
			root: makeMapFS(map[string]string{
				"files/share/applications/com.example.app.desktop":   "[Desktop Entry]\nExec=app\n",
				"files/share/applications/com.example2.desktop":      "[Desktop Entry]\nExec=/abs/path\n",
			}),
			id:      "com.example.app",
			wantErr: true,
		},
		{
			name: "desktop file not matching id",
			root: makeMapFS(map[string]string{
				"files/share/applications/other.app.desktop": "[Desktop Entry]\nExec=app\nIcon=app\n",
			}),
			id:      "com.example.app",
			wantErr: false,
		},
		{
			name: "non-regular file in desktop dir",
			root: func() fs.FS {
				m := make(fstest.MapFS)
				m["files/share/applications/com.example.app.desktop"] = &fstest.MapFile{Data: []byte("[Desktop Entry]\nExec=app\n"), Mode: fs.ModeIrregular}
				return m
			}(),
			id:      "com.example.app",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDesktopFile(tt.root, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDesktopFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDesktopFilename(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		wantErr  bool
	}{
		{"valid dotted name", "com.example.app.desktop", false},
		{"valid single name", "app.desktop", false},
		{"valid with hyphens", "com.example.my-app.desktop", false},
		{"valid with underscores", "com.example.my_app.desktop", false},
		{"valid with numbers", "com.example2.app3.desktop", false},

		{"empty name", ".desktop", true},
		{"empty segment", "com..example.desktop", true},
		{"segment starts with digit", "com.1example.desktop", true},
		{"invalid char $", "com.example$.desktop", true},
		{"invalid char space", "com.example app.desktop", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDesktopFilename(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDesktopFilename(%q) error = %v, wantErr %v", tt.file, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDesktopContent(t *testing.T) {
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
		file    string
		wantErr bool
	}{
		{
			name: "valid relative exec",
			root: makeMapFS(map[string]string{
				"test.desktop": "[Desktop Entry]\nExec=app\nIcon=app\n",
			}),
			file:    "test.desktop",
			wantErr: false,
		},
		{
			name: "absolute exec path",
			root: makeMapFS(map[string]string{
				"test.desktop": "[Desktop Entry]\nExec=/usr/bin/app\n",
			}),
			file:    "test.desktop",
			wantErr: true,
		},
		{
			name: "absolute try exec path",
			root: makeMapFS(map[string]string{
				"test.desktop": "[Desktop Entry]\nTryExec=/opt/app\n",
			}),
			file:    "test.desktop",
			wantErr: true,
		},
		{
			name: "absolute icon path",
			root: makeMapFS(map[string]string{
				"test.desktop": "[Desktop Entry]\nIcon=/usr/share/icons/app.png\n",
			}),
			file:    "test.desktop",
			wantErr: true,
		},
		{
			name: "empty value ignored",
			root: makeMapFS(map[string]string{
				"test.desktop": "[Desktop Entry]\nExec=\n",
			}),
			file:    "test.desktop",
			wantErr: false,
		},
		{
			name: "no executable keys",
			root: makeMapFS(map[string]string{
				"test.desktop": "[Desktop Entry]\nName=Test\nComment=test\n",
			}),
			file:    "test.desktop",
			wantErr: false,
		},
		{
			name: "key without equals ignored",
			root: makeMapFS(map[string]string{
				"test.desktop": "[Desktop Entry]\nExec\n",
			}),
			file:    "test.desktop",
			wantErr: false,
		},
		{
			name: "file open error",
			root: makeMapFS(nil),
			file: "nonexistent.desktop",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDesktopContent(tt.root, tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDesktopContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func FuzzValidateDesktopFilename(f *testing.F) {
	seeds := []string{
		"com.example.app.desktop",
		"app.desktop",
		".desktop",
		"com..example.desktop",
		"com.1example.desktop",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, filename string) {
		validateDesktopFilename(filename)
	})
}
