package gopeed

import "time"

type GopeedStatus string

const (
	GopeedStatusReady   GopeedStatus = "ready"
	GopeedStatusRunning GopeedStatus = "running"
	GopeedStatusPause   GopeedStatus = "pause"
	GopeedStatusWait    GopeedStatus = "wait"
	GopeedStatusError   GopeedStatus = "error"
	GopeedStatusDone    GopeedStatus = "done"
)

type GopeedResponse[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type GopeedInfo struct {
	Version  string `json:"version"`
	Runtime  string `json:"runtime"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	InDocker bool   `json:"inDocker"`
}

type GopeedTask struct {
	ID        string         `json:"id"`
	Protocol  string         `json:"protocol"`
	Name      string         `json:"name"`
	Meta      GopeedMeta     `json:"meta"`
	Status    GopeedStatus   `json:"status"`
	Uploading bool           `json:"uploading"`
	Progress  GopeedProgress `json:"progress"`
	Size      int64          `json:"size"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type GopeedMeta struct {
	Opts GopeedOptions  `json:"opts"`
	Res  GopeedResource `json:"res"`
	Req  GopeedRequest  `json:"req"`
}

type GopeedResource struct {
	Name  string           `json:"name"`
	Size  int64            `json:"size"`
	Range bool             `json:"range"`
	Files []GopeedFileInfo `json:"files"`
	Hash  string           `json:"hash"`
}

type GopeedFileInfo struct {
	Name  string         `json:"name"`
	Path  string         `json:"path"`
	Size  int64          `json:"size"`
	Ctime time.Time      `json:"ctime"`
	Req   *GopeedRequest `json:"req,omitempty"`
}

type GopeedRequest struct {
	URL string `json:"url"`
}

type GopeedProgress struct {
	Used              int64  `json:"used"`
	Speed             int64  `json:"speed"`
	Downloaded        int64  `json:"downloaded"`
	UploadSpeed       int64  `json:"uploadSpeed"`
	Uploaded          int64  `json:"uploaded"`
	ExtractStatus     string `json:"extractStatus"`
	ExtractProgress   int    `json:"extractProgress"`
	MultiPartBaseName string `json:"multiPartBaseName"`
	MultiPartNumber   int    `json:"multiPartNumber"`
	MultiPartIsFirst  bool   `json:"multiPartIsFirst"`
}

type GopeedResolved struct {
	Id       string         `json:"id"`
	Resource GopeedResource `json:"res"`
}

type GopeedResolve struct {
	Req  GopeedRequest `json:"req"`
	Opts GopeedOptions `json:"opts"`
}
type GopeedCreateTask struct {
	Rid  string        `json:"rid"`
	Opts GopeedOptions `json:"opts"`
}
type GopeedExtraOptions struct {
	Connections int `json:"connections"`
}

type GopeedOptions struct {
	Name        string              `json:"name,omitempty"`
	Path        string              `json:"path,omitempty"`
	SelectFiles []int               `json:"selectFiles,omitempty"`
	Extra       *GopeedExtraOptions `json:"extra,omitempty"`
}
