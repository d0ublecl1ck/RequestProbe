package models

// ResourceMonitorSettings 表示资源监听相关的可配置项
type ResourceMonitorSettings struct {
	SaveRootDir        string `json:"saveRootDir"`
	DefaultSaveRootDir string `json:"defaultSaveRootDir"`
}
