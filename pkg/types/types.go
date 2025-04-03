package types

type LayerFileMetaInfo struct {
	Info      LayerInfo `json:"info"`
	Version   string    `json:"version"`
	ErofsSize int64     `json:"erofs_size"`
	SignSize  int64     `json:"sign_size"`

	Head string `json:"head,omitempty"`
	Raw  string `json:"raw,omitempty"`
}

type ApplicationPermission struct{}

type LayerInfo struct {
	ID            string                `json:"id"`
	Appid         string                `json:"appid"`
	Arch          []string              `json:"arch"`
	Base          string                `json:"base"`
	Description   string                `json:"description"`
	Kind          string                `json:"kind"`
	Module        string                `json:"module"`
	Name          string                `json:"name"`
	Runtime       string                `json:"runtime"`
	Size          int64                 `json:"size"`
	Version       string                `json:"version"`
	Channel       string                `json:"channel"`
	SchemaVersion string                `json:"schema_version"`
	Permissions   ApplicationPermission `json:"permissions,omitempty"` // for now, Permissions is not required, so un/marshal to a raw string
}

type UABLayerInfo struct {
	Info     LayerInfo `json:"info"`
	Minified bool      `json:"minified"`
}

type UABFileMetaInfo struct {
	Digest   string            `json:"digest"`
	Layers   []UABLayerInfo    `json:"layers"`
	Sections map[string]string `json:"sections"`
	UUID     string            `json:"uuid"`
	Version  string            `json:"version"`
}
