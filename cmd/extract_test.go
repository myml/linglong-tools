package cmd

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func genData(size int) []byte {
	b := make([]byte, size)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func generateMinimalUAB(root string, bundle string) (string, error) {
	uabPath := filepath.Join(root, "minimal.uab")
	file, err := os.Create(uabPath)
	if err != nil {
		return "", fmt.Errorf("Error creating file: %w", err)
	}
	defer file.Close()

	info, err := os.Stat(bundle)
	if err != nil {
		return "", fmt.Errorf("stat bundle: %w", err)
	}

	var STOffset uint64 = 64

	bundleSize := uint64(info.Size())
	STOffset += bundleSize

	bundleSection := "linglong.bundle\x00"
	shstrtabSection := ".shstrtab\x00"

	STOffset += uint64(len(bundleSection))
	STOffset += uint64(len(shstrtabSection))

	elfHeader := make([]byte, 64)
	elfHeader[0] = 0x7F
	elfHeader[1] = 'E'
	elfHeader[2] = 'L'
	elfHeader[3] = 'F'
	elfHeader[4] = 2 // 64-bit
	elfHeader[5] = 1 // Little-endian
	elfHeader[6] = 1
	elfHeader[7] = 0

	// none type
	binary.LittleEndian.PutUint16(elfHeader[16:], 0)

	// machine type: x86-64
	binary.LittleEndian.PutUint16(elfHeader[18:], 62)

	// ELF version
	binary.LittleEndian.PutUint32(elfHeader[20:], 1)

	// no entry point
	binary.LittleEndian.PutUint64(elfHeader[24:], 0)

	// Program header offset
	binary.LittleEndian.PutUint64(elfHeader[32:], 0)

	// Section header table offset
	binary.LittleEndian.PutUint64(elfHeader[40:], STOffset)

	// No flag
	binary.LittleEndian.PutUint32(elfHeader[48:], 0)

	// ELF header size
	binary.LittleEndian.PutUint16(elfHeader[52:], 64)

	// No program header size
	binary.LittleEndian.PutUint16(elfHeader[54:], 0)

	// No number of program headers
	binary.LittleEndian.PutUint16(elfHeader[56:], 0)

	// Section header size
	binary.LittleEndian.PutUint16(elfHeader[58:], 64) // 64 bit

	// Section header count
	// linglong.bundle and Section header string table
	binary.LittleEndian.PutUint16(elfHeader[60:], 2)

	// Section header string table index
	binary.LittleEndian.PutUint16(elfHeader[62:], 1)

	_, err = file.Write(elfHeader)
	if err != nil {
		return "", fmt.Errorf("Error writing elfHeader: %w", err)
	}

	// Write data of linglong.bundle
	bundleF, err := os.Open(bundle)
	if err != nil {
		return "", fmt.Errorf("Error opening bundle: %w", err)
	}
	defer bundleF.Close()

	buf := make([]byte, 4096)
	for {
		bytes, err := bundleF.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				return "", fmt.Errorf("read failed: %w", err)
			}
			break
		}

		_, err = file.Write(buf[:bytes])
		if err != nil {
			return "", fmt.Errorf("write failed: %w", err)
		}
	}

	// Write data of shstrtab
	_, err = file.WriteString(bundleSection)
	if err != nil {
		return "", fmt.Errorf("failed to write bundle name to shstrtab: %w", err)
	}

	_, err = file.WriteString(shstrtabSection)
	if err != nil {
		return "", fmt.Errorf("failed to write shstrtab name to shstrtab: %w", err)
	}

	// process section header table
	sectionHeader := make([]byte, 128)

	// linglong.bundle
	binary.LittleEndian.PutUint32(sectionHeader[0:], 0)   // name
	binary.LittleEndian.PutUint32(sectionHeader[4:], 1)   // type
	binary.LittleEndian.PutUint64(sectionHeader[8:], 0)   // flag
	binary.LittleEndian.PutUint64(sectionHeader[16:], 0)  // addr
	binary.LittleEndian.PutUint64(sectionHeader[24:], 64) // offset
	binary.LittleEndian.PutUint64(sectionHeader[32:], 64) // size
	binary.LittleEndian.PutUint32(sectionHeader[40:], 0)  // link
	binary.LittleEndian.PutUint32(sectionHeader[44:], 0)  // info
	binary.LittleEndian.PutUint64(sectionHeader[48:], 1)  // addralign
	binary.LittleEndian.PutUint64(sectionHeader[56:], 0)  // entsize

	// Section header string table
	strTableSize := uint64(len(bundleSection) + len(shstrtabSection))
	binary.LittleEndian.PutUint32(sectionHeader[64:], uint32(len(bundleSection)))
	binary.LittleEndian.PutUint32(sectionHeader[68:], 3)
	binary.LittleEndian.PutUint64(sectionHeader[72:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[80:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[88:], 64+bundleSize)
	binary.LittleEndian.PutUint64(sectionHeader[96:], strTableSize)
	binary.LittleEndian.PutUint32(sectionHeader[104:], 0)
	binary.LittleEndian.PutUint32(sectionHeader[108:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[112:], 1)
	binary.LittleEndian.PutUint64(sectionHeader[120:], 0)

	_, err = file.Write(sectionHeader)
	if err != nil {
		return "", fmt.Errorf("failed to write section header table: %w", err)
	}

	return uabPath, nil
}

func generateMinimalBundle(root string) (string, error) {
	bundleDir := filepath.Join(root, "bundle")
	err := os.MkdirAll(bundleDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", bundleDir, err)
	}

	// binary
	data := genData(1024 * 1024)
	binPath := filepath.Join(bundleDir, "fake-loader")
	err = os.WriteFile(binPath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create binary %s: %w", binPath, err)
	}

	// sub directory
	extraDir := filepath.Join(bundleDir, "subDir")
	err = os.MkdirAll(extraDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", extraDir, err)
	}

	// symlink
	symName := filepath.Join(bundleDir, "subDir-symlink")
	err = os.Symlink("extra", symName)
	if err != nil {
		return "", fmt.Errorf("failed to create symlink %s: %w", symName, err)
	}

	// regular file in subDir
	for i := 0; i < 10; i++ {
		tmpData := genData(1024 * 256)
		tmpName := "test-file-" + string(tmpData[:4]) + ".txt"
		tmpFile := filepath.Join(extraDir, tmpName)
		err = os.WriteFile(tmpFile, data, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to writeFile %s: %w", tmpFile, err)
		}

		if i%2 == 0 {
			symName := filepath.Join(extraDir, "sym-"+tmpName)
			err = os.Symlink(tmpFile, symName)
			if err != nil {
				return "", fmt.Errorf("failed to create symlink %s: %w", symName, err)
			}
		}
	}

	bundleFile := filepath.Join(root, "bundle.ef")
	cmd := exec.Command("mkfs.erofs", "-zlz4hc,9", "-b4096", bundleFile, bundleDir)
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed make erofs image: %w", err)
	}

	return bundleFile, nil
}

func TestExtractRun(t *testing.T) {
	assert := require.New(t)

	// test extract uab
	_, err := exec.LookPath("mkfs.erofs")
	assert.NoError(err)

	_, err = exec.LookPath("objcopy")
	assert.NoError(err)
	temp, err := os.MkdirTemp("", "test-uabBundle-*")
	assert.NoError(err)
	defer os.RemoveAll(temp)

	bundleFile, err := generateMinimalBundle(temp)
	assert.NoError(err)

	uab, err := generateMinimalUAB(temp, bundleFile)
	assert.NoError(err)

	args := ExtractArgs{
		File:      uab,
		OutputDir: filepath.Join(temp, "output"),
	}

	assert.NoError(extractRun(args))
}
