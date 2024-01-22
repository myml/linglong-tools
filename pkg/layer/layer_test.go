package layer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type BinMetaInfo struct {
	Head [40]byte
	Size uint32
}

func TestParseMetaInfoMock(t *testing.T) {
	assert := require.New(t)
	appID := "test"
	head := "<<< linglong >>>"
	var metaInfo MetaInfo
	metaInfo.Info.Appid = appID
	payload, err := json.Marshal(metaInfo)
	assert.NoError(err)

	var buff bytes.Buffer
	// test empty read
	tmpBuf := bytes.NewBuffer(buff.Bytes())
	_, err = ParseMetaInfo(tmpBuf)
	assert.Error(err)

	buff.WriteString(head)
	buff.Write(bytes.Repeat([]byte{0}, 40))

	// test without size
	tmpBuf = bytes.NewBuffer(buff.Bytes()[:40])
	_, err = ParseMetaInfo(tmpBuf)
	assert.Error(err)

	rawInfo := BinMetaInfo{
		Head: [40]byte(buff.Bytes()[:40]),
		Size: uint32(len(payload)),
	}

	buff.Reset()
	err = binary.Write(&buff, binary.LittleEndian, rawInfo)
	assert.NoError(err)
	// test without payload
	tmpBuf = bytes.NewBuffer(buff.Bytes())
	_, err = ParseMetaInfo(tmpBuf)
	assert.Error(err)
	// test invalid payload
	tmpBuf = bytes.NewBuffer(buff.Bytes())
	tmpBuf.Write(bytes.Repeat([]byte{0}, len(payload)))
	_, err = ParseMetaInfo(tmpBuf)
	assert.Error(err)

	_, err = buff.Write(payload)
	assert.NoError(err)
	info, err := ParseMetaInfo(&buff)
	assert.NoError(err)
	assert.Equalf([]byte(info.Info.Appid), []byte(metaInfo.Info.Appid), "raw: %s parse: %#v", payload, info)
}

func TestParseMetaInfoReal(t *testing.T) {
	assert := require.New(t)
	data, err := os.ReadFile("testdata/test.layer")
	assert.NoError(err)
	info, err := ParseMetaInfo(bytes.NewReader(data))
	assert.NoError(err)
	assert.Equal(info.Raw, `{"info":{"appid":"org.deepin.draw","arch":["x86_64"],"base":"/latest/x86_64","description":"draw for deepin os.\n","kind":"app","module":"runtime","name":"deepin-draw","runtime":"org.deepin.Runtime/23.0.0/x86_64","size":102887829,"version":"6.0.5"},"version":"0.1"}`)
}
