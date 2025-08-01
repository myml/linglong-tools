package tarutils

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTar(t *testing.T) {
	assert := require.New(t)
	// 测试创建tar包
	tarFile := "1.tar"
	{
		f, err := os.Create(tarFile)
		assert.NoError(err)
		err = CreateTar(f, "../../cmd")
		assert.NoError(err)
	}
	// 测试解压tar包
	{
		tmp, err := os.MkdirTemp("", "")
		assert.NoError(err)
		f, err := os.Open(tarFile)
		assert.NoError(err)
		err = ExtractTar(f, tmp)
		assert.NoError(err)
		assert.NoError(os.RemoveAll(tmp))
	}

	assert.NoError(os.Remove(tarFile))
}

func TestTarStream(t *testing.T) {
	assert := require.New(t)
	// 测试创建tar包
	tarFile := "1.tar"
	f, err := os.Create(tarFile)
	assert.NoError(err)
	tarStream, err := CreateTarStream("../../cmd")
	assert.NoError(err)
	_, err = io.Copy(f, tarStream)
	assert.NoError(err)
	assert.NoError(f.Close())
	assert.NoError(os.Remove(tarFile))
}
