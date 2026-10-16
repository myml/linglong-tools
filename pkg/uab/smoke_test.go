package uab

import (
	"debug/elf"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type smokeOptions struct {
	runnable bool
	native   bool
}

func TestSmokeLinkersAndBinutils(t *testing.T) {
	if testing.Short() {
		t.Skip("smoke test skipped in -short")
	}

	cases := []struct {
		name     string
		runnable bool
		build    func(dir string) (string, bool)
	}{
		{"go", true, func(dir string) (string, bool) { return buildGoELF(t, dir) }},
		{"cc-ld", true, func(dir string) (string, bool) { return buildCELExec(t, dir, "cc", "cc-bin") }},
		{"clang-lld", true, func(dir string) (string, bool) {
			return buildCELExec(t, dir, "clang", "clang-lld-bin", "-fuse-ld=lld")
		}},
		{"mold", true, func(dir string) (string, bool) {
			if _, err := exec.LookPath("mold"); err != nil {
				return "", false
			}
			return buildCELExec(t, dir, "cc", "mold-bin", "-fuse-ld=mold")
		}},
		{"shared-lib", false, func(dir string) (string, bool) {
			return buildC(t, dir, "cc", "libsmoke.so", "int smoke(void){return 0;}\n", "-shared", "-fPIC")
		}},
		{"relocatable", false, func(dir string) (string, bool) {
			return buildC(t, dir, "cc", "smoke.o", "int smoke(void){return 0;}\n", "-c")
		}},
		{"system-true", true, func(dir string) (string, bool) { return copySystemELF(t, dir, "/bin/true") }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path, ok := tc.build(t.TempDir())
			if !ok {
				t.Skip("toolchain or input unavailable")
			}

			if !isELF64(t, path) {
				t.Skip("not a 64-bit ELF")
			}

			smokeSignLifecycle(t, path, smokeOptions{runnable: tc.runnable, native: true})
		})
	}
}

func TestSmokeCrossArchitectures(t *testing.T) {
	if testing.Short() {
		t.Skip("smoke test skipped in -short")
	}

	if !haveTool("go") {
		t.Skip("go not found")
	}

	for _, goarch := range []string{"arm64", "riscv64", "mips64", "loong64"} {
		t.Run(goarch, func(t *testing.T) {
			path, ok := buildGoELFForArch(t, t.TempDir(), goarch)
			if !ok {
				t.Skip("cross build unavailable")
			}

			if !isELF64(t, path) {
				t.Skip("not a 64-bit ELF")
			}

			smokeSignLifecycle(t, path, smokeOptions{runnable: false, native: false})
		})
	}
}

func smokeSignLifecycle(t *testing.T, path string, opts smokeOptions) {
	t.Helper()

	verify := func(stage, payload string) {
		requireSectionData(t, path, map[string][]byte{signSectionName: []byte(payload)})
		if opts.runnable {
			requireRun(t, path, stage)
		}

		requireReadelfClean(t, path, stage)
		requireObjdumpClean(t, path, stage)
		requireToolRoundTrip(t, path, stage, payload, opts)
	}

	signAction(t, path, false, "signature-A")
	verify("add", "signature-A")

	signAction(t, path, true, "signature-BB-longer")
	verify("grow", "signature-BB-longer")

	signAction(t, path, true, "s")
	verify("shrink", "s")
}

func haveTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func isELF64(t *testing.T, path string) bool {
	t.Helper()
	ef, err := elf.Open(path)
	if err != nil {
		return false
	}

	defer func() { _ = ef.Close() }()
	return ef.Class == elf.ELFCLASS64
}

func requireRun(t *testing.T, path, stage string) {
	t.Helper()
	out, err := exec.Command(path).CombinedOutput()
	require.NoErrorf(t, err, "%s: running %s failed: %s", stage, path, out)
}

func requireReadelfClean(t *testing.T, path, stage string) {
	t.Helper()
	if !haveTool("readelf") {
		return
	}

	out, err := exec.Command("readelf", "-h", "-S", "-W", "-l", path).CombinedOutput()
	require.NoErrorf(t, err, "%s: readelf rejected the file: %s", stage, out)
	require.Containsf(t, string(out), signSectionName,
		"%s: readelf did not list section %q", stage, signSectionName)
}

func requireObjdumpClean(t *testing.T, path, stage string) {
	t.Helper()
	if !haveTool("objdump") {
		return
	}

	out, err := exec.Command("objdump", "-h", path).CombinedOutput()
	require.NoErrorf(t, err, "%s: objdump rejected the file: %s", stage, out)
}

func requireToolRoundTrip(t *testing.T, path, stage, payload string, opts smokeOptions) {
	t.Helper()

	roundTrip := func(tool, out string, args ...string) {
		t.Helper()
		if !haveTool(tool) {
			return
		}

		runTool(t, stage, tool, args...)
		requireSectionData(t, out, map[string][]byte{signSectionName: []byte(payload)})
		if opts.runnable {
			requireRun(t, out, stage+"-"+tool)
		}

		_ = os.Remove(out)
	}

	if opts.native {
		objcopyOut := path + "." + stage + ".objcopy"
		roundTrip("objcopy", objcopyOut, path, objcopyOut)

		stripOut := path + "." + stage + ".strip"
		roundTrip("strip", stripOut, "--strip-debug", path, "-o", stripOut)
	}

	llvmObjcopyOut := path + "." + stage + ".llvmobjcopy"
	roundTrip("llvm-objcopy", llvmObjcopyOut, path, llvmObjcopyOut)

	llvmStripOut := path + "." + stage + ".llvmstrip"
	roundTrip("llvm-strip", llvmStripOut, "--strip-debug", path, "-o", llvmStripOut)
}

func runTool(t *testing.T, stage, name string, args ...string) {
	t.Helper()
	out, err := exec.Command(name, args...).CombinedOutput()
	require.NoErrorf(t, err, "%s: %s failed: %s", stage, name, out)
}

func buildGoELF(t *testing.T, dir string) (string, bool) {
	t.Helper()
	return buildGoELFForArch(t, dir, runtime.GOARCH)
}

func buildGoELFForArch(t *testing.T, dir, goarch string) (string, bool) {
	t.Helper()
	if !haveTool("go") {
		t.Skip("go not found")
	}

	srcDir := filepath.Join(dir, "gosrc")
	require.NoError(t, os.MkdirAll(srcDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "go.mod"), []byte("module smoke\n\ngo 1.22\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644))

	out := filepath.Join(dir, "go-"+goarch)
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH="+goarch)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Logf("go build %s failed: %v\n%s", goarch, err, b)
		return "", false
	}

	return out, true
}

func buildCELExec(t *testing.T, dir, cc, out string, extra ...string) (string, bool) {
	t.Helper()
	return buildC(t, dir, cc, out, "int main(void){return 0;}\n", extra...)
}

func buildC(t *testing.T, dir, cc, out, src string, args ...string) (string, bool) {
	t.Helper()
	if !haveTool(cc) {
		return "", false
	}

	srcPath := filepath.Join(dir, "smoke.c")
	require.NoError(t, os.WriteFile(srcPath, []byte(src), 0o644))
	outPath := filepath.Join(dir, out)

	full := append(append([]string{}, args...), "-o", outPath, srcPath)
	if b, err := exec.Command(cc, full...).CombinedOutput(); err != nil {
		t.Logf("%s build failed: %v\n%s", cc, err, b)
		return "", false
	}

	return outPath, true
}

func copySystemELF(t *testing.T, dir, src string) (string, bool) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		return "", false
	}

	dst := filepath.Join(dir, filepath.Base(src))
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatalf("copy %s: %v", src, err)
	}

	return dst, true
}
