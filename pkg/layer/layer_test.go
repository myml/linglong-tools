package layer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/myml/linglong-tools/pkg/types"
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
	var metaInfo types.LayerFileMetaInfo
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

func TestLayer(t *testing.T) {
	assert := require.New(t)

	workdir, err := os.MkdirTemp("", "test")
	assert.NoError(err)
	t.Log(workdir)
	defer os.RemoveAll(workdir)

	f, err := os.Create(filepath.Join(workdir, "test.layer"))
	assert.NoError(err)
	// 将头信息写入
	data, err := os.ReadFile("testdata/test.layer")
	assert.NoError(err)
	headerLen := len(data)
	_, err = f.Write(data)
	assert.NoError(err)
	// 制作一个erofs文件，并写入到layer文件中
	_, err = exec.Command("mkfs.erofs", filepath.Join(workdir, "test.erofs"), "../../cmd").CombinedOutput()
	assert.NoError(err)
	erofs, err := os.ReadFile(filepath.Join(workdir, "test.erofs"))
	assert.NoError(err)
	_, err = f.Write(erofs)
	assert.NoError(err)
	assert.NoError(f.Close())
	t.Log(f.Name())

	// 插入签名
	{
		signDir := filepath.Join(workdir, "sign")
		os.Mkdir(signDir, 0755)
		os.WriteFile(filepath.Join(signDir, "test.txt"), []byte("test"), 0644)
		layer, err := NewLayer(filepath.Join(workdir, "test.layer"))
		assert.NoError(err)
		assert.Equal(layer.erofsOffset, int64(headerLen))
		err = layer.InsertSign(filepath.Join(workdir, "signed.layer"), signDir)
		assert.NoError(err)
		layer, err = NewLayer(filepath.Join(workdir, "signed.layer"))
		assert.NoError(err)
		assert.True(layer.HasSign())
		assert.True(layer.meta.SignSize > 0)
		assert.Equal(layer.meta.ErofsSize, uint64(len(erofs)))
	}

	// 读取layer文件，并验证内容
	layer, err := NewLayer(filepath.Join(workdir, "signed.layer"))
	assert.NoError(err)
	assert.Equal(layer.meta.Info.Appid, "org.deepin.draw")
	assert.Equal(layer.meta.Info.Arch, []string{"x86_64"})
	assert.Equal(layer.meta.Info.Base, "/latest/x86_64")
	assert.Equal(layer.meta.Info.Description, "draw for deepin os.\n")
	assert.Equal(layer.meta.Info.Kind, "app")
	assert.Equal(layer.meta.Info.Module, "runtime")
	assert.True(layer.erofsOffset > int64(headerLen))
	err = layer.SaveErofs(filepath.Join(workdir, "test2.erofs"))
	assert.NoError(err)
	// 提取erofs文件
	extractDir := filepath.Join(workdir, "extract")
	os.Mkdir(extractDir, 0755)
	assert.NoError(layer.Extract(extractDir))
	if layer.HasSign() {
		// 提取签名数据
		extractDir = filepath.Join(workdir, "extract-sign")
		os.Mkdir(extractDir, 0755)
		assert.NoError(layer.ExtractSign(extractDir))
	}
}
