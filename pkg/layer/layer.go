package layer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// ParseMetaInfo parse layer metainfo
// layer file format <head(chars 40 byte)> <json payload size(uint32 4 byte)> <json payload(varchar)> <erofs image>
func ParseMetaInfo(r io.Reader) (*MetaInfo, error) {
	var buff bytes.Buffer
	_, err := io.CopyN(&buff, r, int64(40))
	if err != nil {
		return nil, fmt.Errorf("read head: %w", err)
	}
	head := buff.String()
	buff.Reset()
	var size uint32
	err = binary.Read(r, binary.LittleEndian, &size)
	if err != nil {
		return nil, fmt.Errorf("read info size: %w", err)
	}
	_, err = io.CopyN(&buff, r, int64(size))
	if err != nil {
		return nil, fmt.Errorf("read info data: %w", err)
	}
	var info MetaInfo
	err = json.Unmarshal(buff.Bytes(), &info)
	if err != nil {
		return nil, fmt.Errorf("unmarshal info data: %w", err)
	}
	info.Head = head
	info.Raw = buff.String()
	return &info, nil
}

type MetaInfo struct {
	Head    string  `json:"head"`
	Info    AppInfo `json:"info"`
	Version string  `json:"version"`
	Raw     string  `json:"raw"`
}

type AppInfo struct {
	Appid       string   `json:"appid"`
	Arch        []string `json:"arch"`
	Base        string   `json:"base"`
	Description string   `json:"description"`
	Kind        string   `json:"kind"`
	Module      string   `json:"module"`
	Name        string   `json:"name"`
	Runtime     string   `json:"runtime"`
	Size        int      `json:"size"`
	Version     string   `json:"version"`
}
