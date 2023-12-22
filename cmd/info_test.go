package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/stretchr/testify/require"
)

// 生成一个临时layer文件用于测试
func genLayerFile(assert *require.Assertions, info layer.MetaInfo) string {
	// 写入头部标识
	var buff bytes.Buffer
	buff.WriteString(info.Head)
	buff.Write(bytes.Repeat([]byte{0}, 40))
	buff.Truncate(40)

	payload, err := json.Marshal(info)
	assert.NoError(err)
	// 写入payload size
	err = binary.Write(&buff, binary.LittleEndian, uint32(len(payload)))
	assert.NoError(err)
	// 写入payload
	_, err = buff.Write(payload)
	assert.NoError(err)
	// 写入erofs内容
	_, err = buff.WriteString("erofs image content")
	assert.NoError(err)
	// 创建临时文件
	f, err := os.CreateTemp("", "")
	assert.NoError(err)
	defer f.Close()
	// 将缓存区写入临时文件中
	_, err = f.Write(buff.Bytes())
	assert.NoError(err)
	err = f.Close()
	assert.NoError(err)
	return f.Name()
}

func TestInfoRun(t *testing.T) {
	assert := require.New(t)
	head := "<<< linglong >>>"
	appID := "test"

	// 生成文件
	var metaInfo layer.MetaInfo
	metaInfo.Head = head
	metaInfo.Info.Appid = appID
	metaInfo.Info.Arch = append(metaInfo.Info.Arch, "amd64")
	fname := genLayerFile(assert, metaInfo)
	// 测试file参数
	infoArgs.LayerFile = fname
	assert.NoError(InfoRun(infoArgs))
	// 测试prettier参数
	infoArgs.PrettierOutput = true
	assert.NoError(InfoRun(infoArgs))
	// 测试format参数
	infoArgs.FormatOutput = "{{ .Raw }}"
	assert.NoError(InfoRun(infoArgs))
	infoArgs.FormatOutput = "{{ index .Info.Arch 0 }}"
	assert.NoError(InfoRun(infoArgs))
	infoCmd.Run(nil, nil)
	// 测试format数组越界
	infoArgs.FormatOutput = "{{ index .Info.Arch 1 }}"
	assert.Error(InfoRun(infoArgs))
}
