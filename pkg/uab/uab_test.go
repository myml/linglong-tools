package uab

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const testMetaJson = `{
    "digest": "7ffaf77c13d8fe94f824a9974e5f30911429df52825ba624a3b70b91b8f67876",
    "layers": [
        {
            "info": {
                "arch": [
                    "x86_64"
                ],
                "base": "org.deepin.foundation/20.0.2",
                "channel": "main",
                "description": "deepin base environment.\n",
                "id": "org.deepin.base",
                "kind": "runtime",
                "module": "binary",
                "name": "deepin-foundation",
                "permissions": {},
                "runtime": "latest",
                "schema_version": "1.0",
                "size": 413747564,
                "version": "23.1.0.3"
            },
            "minified": false
        },
        {
            "info": {
                "arch": [
                    "x86_64"
                ],
                "base": "main:org.deepin.base/23.1.0/x86_64",
                "channel": "main",
                "command": [
                    "echo",
                    "hello, world"
                ],
                "description": "your description #set a brief text to introduce your application.\n",
                "id": "org.deepin.demo",
                "kind": "app",
                "module": "binary",
                "name": "your name",
                "schema_version": "1.0",
                "size": 8192,
                "version": "0.0.0.1"
            },
            "minified": false
        }
    ],
    "sections": {
        "bundle": "linglong.bundle"
    },
    "uuid": "b2e7aab7-290d-44ef-b94d-9a04a49a25f4",
    "version": "1"
}`

type UABTestSuite struct {
	TmpUAB string
	suite.Suite
}

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
	metaSize := uint64(len(testMetaJson))
	STOffset += (bundleSize + metaSize)

	bundleSection := "linglong.bundle\x00"
	metaSection := "linglong.meta\x00"
	shstrtabSection := ".shstrtab\x00"
	shstrtabSectionSize := uint64(len(bundleSection) + len(metaSection) + len(shstrtabSection))

	STOffset += shstrtabSectionSize

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
	// linglong.bundle, linglong.meta and Section header string table
	binary.LittleEndian.PutUint16(elfHeader[60:], 3)

	// Section header string table index
	binary.LittleEndian.PutUint16(elfHeader[62:], 2)

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
	bundleFInfo, err := bundleF.Stat()
	if err != nil {
		return "", fmt.Errorf("stat bundle file: %w", err)
	}
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

	// Write data of linglong.meta
	_, err = file.WriteString(testMetaJson)
	if err != nil {
		return "", fmt.Errorf("failed to write meta data to uab: %w", err)
	}

	// Write data of shstrtab
	_, err = file.WriteString(bundleSection)
	if err != nil {
		return "", fmt.Errorf("failed to write bundle name to shstrtab: %w", err)
	}

	_, err = file.WriteString(metaSection)
	if err != nil {
		return "", fmt.Errorf("failed to write meta name to shstrtab: %w", err)
	}

	_, err = file.WriteString(shstrtabSection)
	if err != nil {
		return "", fmt.Errorf("failed to write shstrtab name to shstrtab: %w", err)
	}

	// process section header table
	sectionHeader := make([]byte, 192)
	sectionNameOffset := 0
	// linglong.bundle
	binary.LittleEndian.PutUint32(sectionHeader[0:], uint32(sectionNameOffset))   // name
	binary.LittleEndian.PutUint32(sectionHeader[4:], 1)                           // type
	binary.LittleEndian.PutUint64(sectionHeader[8:], 0)                           // flag
	binary.LittleEndian.PutUint64(sectionHeader[16:], 0)                          // addr
	binary.LittleEndian.PutUint64(sectionHeader[24:], 64)                         // offset
	binary.LittleEndian.PutUint64(sectionHeader[32:], uint64(bundleFInfo.Size())) // size
	binary.LittleEndian.PutUint32(sectionHeader[40:], 0)                          // link
	binary.LittleEndian.PutUint32(sectionHeader[44:], 0)                          // info
	binary.LittleEndian.PutUint64(sectionHeader[48:], 1)                          // addralign
	binary.LittleEndian.PutUint64(sectionHeader[56:], 0)                          // entsize

	//linglong.meta
	sectionNameOffset += len(bundleSection)
	binary.LittleEndian.PutUint32(sectionHeader[64:], uint32(sectionNameOffset))
	binary.LittleEndian.PutUint32(sectionHeader[68:], 1)
	binary.LittleEndian.PutUint64(sectionHeader[72:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[80:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[88:], 64+bundleSize)
	binary.LittleEndian.PutUint64(sectionHeader[96:], uint64(len(testMetaJson)))
	binary.LittleEndian.PutUint32(sectionHeader[104:], 0)
	binary.LittleEndian.PutUint32(sectionHeader[108:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[112:], 1)
	binary.LittleEndian.PutUint64(sectionHeader[120:], 0)

	// Section header string table
	sectionNameOffset += len(metaSection)
	binary.LittleEndian.PutUint32(sectionHeader[128:], uint32(sectionNameOffset))
	binary.LittleEndian.PutUint32(sectionHeader[132:], 3)
	binary.LittleEndian.PutUint64(sectionHeader[136:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[144:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[152:], 64+bundleSize+metaSize)
	binary.LittleEndian.PutUint64(sectionHeader[160:], shstrtabSectionSize)
	binary.LittleEndian.PutUint32(sectionHeader[168:], 0)
	binary.LittleEndian.PutUint32(sectionHeader[172:], 0)
	binary.LittleEndian.PutUint64(sectionHeader[176:], 1)
	binary.LittleEndian.PutUint64(sectionHeader[184:], 0)

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

func (suite *UABTestSuite) TestExtract() {
	assert := suite.Require()

	uabFile, err := Open(suite.TmpUAB)
	assert.NoError(err)
	defer uabFile.Close()

	extractDir, err := ioutil.TempDir("", "test-extract-*")
	assert.NoError(err)
	defer os.RemoveAll(extractDir)

	err = uabFile.Extract(extractDir)
	assert.NoError(err)
}

func TestUABTestSuite(t *testing.T) {
	_, err := exec.LookPath("mkfs.erofs")
	if err != nil {
		t.Skip("mkfs.erofs not found")
	}

	_, err = exec.LookPath("objcopy")
	if err != nil {
		t.Skip("objcopy not found")
	}

	temp, err := ioutil.TempDir("", "test-uabBundle-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err.Error())
	}
	defer os.RemoveAll(temp)

	bundleFile, err := generateMinimalBundle(temp)
	if err != nil {
		t.Fatalf("failed to generate minimal bundle: %s", err.Error())
	}

	testUAB, err := generateMinimalUAB(temp, bundleFile)
	if err != nil {
		t.Fatalf("failed to generate minimal uab: %s", err.Error())
	}

	uabSuite := new(UABTestSuite)
	uabSuite.TmpUAB = testUAB

	suite.Run(t, uabSuite)
}
