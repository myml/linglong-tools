package tarutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTar(t *testing.T) {
	assert := require.New(t)
	f, err := os.Create("1.tar")
	assert.NoError(err)
	err = CreateTar(f, "../../tools")
	assert.NoError(err)
	os.Remove("1.tar")
}
