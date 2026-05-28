package checker

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestCheckSystemd(t *testing.T) {
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
			name:    "no systemd dir",
			root:    makeMapFS(nil),
			wantErr: false,
		},
		{
			name: "valid service with relative exec",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/myapp.service": "[Service]\nExecStart=myapp\n",
				"files/bin/myapp":                 "binary",
			}),
			wantErr: false,
		},
		{
			name: "unsupported unit type",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/myapp.conf": "[Service]\n",
			}),
			wantErr: true,
		},
		{
			name: "service with absolute ExecStart",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/myapp.service": "[Service]\nExecStart=/usr/bin/myapp\n",
			}),
			wantErr: true,
		},
		{
			name: "service with missing binary",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/myapp.service": "[Service]\nExecStart=myapp\n",
			}),
			wantErr: true,
		},
		{
			name: "valid socket unit",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/myapp.socket": "[Socket]\nExecStart=sock-handler\n",
				"files/bin/sock-handler":         "binary",
			}),
			wantErr: false,
		},
		{
			name: "invalid filename starts with digit",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/1myapp.service": "[Service]\n",
			}),
			wantErr: true,
		},
		{
			name: "valid timer unit (no content check)",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/myapp.timer": "[Timer]\nOnBootSec=5min\n",
			}),
			wantErr: false,
		},
		{
			name: "ExecStart with arguments",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/myapp.service": "[Service]\nExecStart=myapp --config /etc/config\n",
				"files/bin/myapp":                 "binary",
			}),
			wantErr: false,
		},
		{
			name: "ExecStart= empty ignored",
			root: makeMapFS(map[string]string{
				"files/lib/systemd/myapp.service": "[Service]\nExecStart=\n",
			}),
			wantErr: false,
		},
		{
			name: "empty systemd dir",
			root: func() fs.FS {
				m := make(fstest.MapFS)
				m["files/lib/systemd"] = &fstest.MapFile{Mode: fs.ModeDir}
				return m
			}(),
			wantErr: false,
		},
		{
			name: "non-regular file in systemd dir",
			root: func() fs.FS {
				m := make(fstest.MapFS)
				m["files/lib/systemd/myapp.service"] = &fstest.MapFile{Data: []byte("[Service]\n"), Mode: fs.ModeIrregular}
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSystemd(tt.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSystemd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateServiceFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"valid service", "myapp.service", false},
		{"valid with hyphen", "my-app.service", false},
		{"valid with underscore", "my_app.service", false},
		{"valid with @", "myapp@.service", false},
		{"valid socket", "myapp.socket", false},
		{"valid timer", "myapp.timer", false},
		{"valid multi-dot", "my.app.service", false},

		{"empty name", ".service", true},
		{"starts with digit", "1myapp.service", true},
		{"starts with hyphen", "-myapp.service", true},
		{"invalid char $", "my$app.service", true},
		{"invalid char space", "my app.service", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateServiceFilename(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			}
		})
	}
}

func TestValidateServiceContent(t *testing.T) {
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
				"test.service": "[Service]\nExecStart=myapp\n",
				"files/bin/myapp": "binary",
			}),
			file:    "test.service",
			wantErr: false,
		},
		{
			name: "absolute exec path",
			root: makeMapFS(map[string]string{
				"test.service": "[Service]\nExecStart=/usr/bin/myapp\n",
			}),
			file:    "test.service",
			wantErr: true,
		},
		{
			name: "missing binary",
			root: makeMapFS(map[string]string{
				"test.service": "[Service]\nExecStart=missing-bin\n",
			}),
			file:    "test.service",
			wantErr: true,
		},
		{
			name: "ExecStartPre with absolute path",
			root: makeMapFS(map[string]string{
				"test.service": "[Service]\nExecStartPre=/bin/prep\n",
			}),
			file:    "test.service",
			wantErr: true,
		},
		{
			name: "empty ExecStart ignored",
			root: makeMapFS(map[string]string{
				"test.service": "[Service]\nExecStart=\n",
			}),
			file:    "test.service",
			wantErr: false,
		},
		{
			name: "file not found",
			root: makeMapFS(nil),
			file:    "nonexistent.service",
			wantErr: true,
		},
		{
			name: "scanner error on long line",
			root: makeMapFS(map[string]string{
				"test.service": "ExecStart=" + string(make([]byte, 70000)),
			}),
			file:    "test.service",
			wantErr: true,
		},
		{
			name: "Exec line without equals sign",
			root: makeMapFS(map[string]string{
				"test.service": "[Service]\nExecStart\n",
			}),
			file:    "test.service",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceContent(tt.root, tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateServiceContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseExecPath(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"myapp", "myapp"},
		{"/usr/bin/myapp", "/usr/bin/myapp"},
		{"myapp --config /etc/cfg", "myapp"},
		{"myapp\t--flag", "myapp"},
		{"", ""},
		{"   ", ""},
		{"/opt/bin/app --arg", "/opt/bin/app"},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got := parseExecPath(tt.value)
			if got != tt.want {
				t.Errorf("parseExecPath(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsBinaryExists(t *testing.T) {
	root := fstest.MapFS{
		"files/bin/myapp": &fstest.MapFile{Data: []byte("binary")},
	}

	tests := []struct {
		binary string
		exists bool
	}{
		{"myapp", true},
		{"nonexistent", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.binary, func(t *testing.T) {
			got := isBinaryExists(root, tt.binary)
			if got != tt.exists {
				t.Errorf("isBinaryExists(%q) = %v, want %v", tt.binary, got, tt.exists)
			}
		})
	}
}

func FuzzValidateServiceFilename(f *testing.F) {
	seeds := []string{
		"myapp.service",
		"my-app@.socket",
		"1bad.service",
		".service",
		"valid.timer",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, filename string) {
		validateServiceFilename(filename)
	})
}

func FuzzParseExecPath(f *testing.F) {
	seeds := []string{
		"myapp",
		"/usr/bin/myapp",
		"myapp --flag arg",
		"",
		"  spaced  ",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		result := parseExecPath(value)
		if result != "" && value != "" {
			if len(result) > len(value) {
				t.Error("result should not be longer than input")
			}
		}
	})
}
