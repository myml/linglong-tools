package erofs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErofs(t *testing.T) {
	assert := require.New(t)
	erofs, err := NewErofsCmd()
	assert.NoError(err)
	workdir, err := os.MkdirTemp("", "test")
	assert.NoError(err)
	defer os.RemoveAll(workdir)
	erofsFile := filepath.Join(workdir, "test.erofs")
	err = erofs.Mkfs(erofsFile, "../../cmd")
	assert.NoError(err)

	offsetFile := filepath.Join(workdir, "test.offset.erofs")
	f, err := os.Create(offsetFile)
	assert.NoError(err)
	f.Write(bytes.Repeat([]byte{0}, 1024))
	data, err := os.ReadFile(erofsFile)
	assert.NoError(err)
	f.Write(data)
	assert.NoError(f.Close())

	erofsFile = offsetFile

	info, err := os.Stat(erofsFile)
	assert.NoError(err)
	err = erofs.Extract(erofsFile, filepath.Join(workdir, "extract"), 1024, info.Size()-1024)
	assert.NoError(err)
	erofs.SupportOffset = false
	err = erofs.Extract(erofsFile, filepath.Join(workdir, "extract2"), 1024, info.Size()-1024)
	assert.NoError(err)

	erofs.SupportOffset = false
	err = erofs.Extract(erofsFile, filepath.Join(workdir, "extract3"), 1024, 0)
	assert.NoError(err)

}
