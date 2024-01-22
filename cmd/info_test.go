package cmd

import (
	"testing"

	"github.com/myml/linglong-tools/pkg/layer"
	"github.com/stretchr/testify/require"
)

// 生成一个临时layer文件用于测试
func genLayerFile(assert *require.Assertions, info layer.MetaInfo) string {
	return "../pkg/layer/testdata/test.layer"
}

func TestInfoRun(t *testing.T) {
	assert := require.New(t)
	head := "<<< linglong >>>"
	appID := "test"
	infoCmd := initInfoCmd()
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
