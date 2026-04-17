package models

// ResourceMonitorStatus 表示资源监听任务状态
type ResourceMonitorStatus string

const (
	ResourceMonitorStatusIdle    ResourceMonitorStatus = "idle"
	ResourceMonitorStatusRunning ResourceMonitorStatus = "running"
	ResourceMonitorStatusPaused  ResourceMonitorStatus = "paused"
	ResourceMonitorStatusEnded   ResourceMonitorStatus = "ended"
	ResourceMonitorStatusError   ResourceMonitorStatus = "error"
)

// ResourceMonitorTask 表示当前资源监听任务
type ResourceMonitorTask struct {
	TaskID             string                `json:"taskId"`
	URL                string                `json:"url"`
	Status             ResourceMonitorStatus `json:"status"`
	SelectedExtensions []string              `json:"selectedExtensions"`
	DownloadDir        string                `json:"downloadDir"`
	CreatedAt          string                `json:"createdAt"`
	UpdatedAt          string                `json:"updatedAt"`
	LastError          string                `json:"lastError,omitempty"`
	Resources          []*MonitoredResource  `json:"resources"`
	Requests           []*MonitoredRequest   `json:"requests"`
}

// MonitoredResource 表示监听到的资源
type MonitoredResource struct {
	ID                string `json:"id"`
	URL               string `json:"url"`
	Extension         string `json:"extension"`
	Hash              string `json:"hash"`
	MimeType          string `json:"mimeType,omitempty"`
	StatusCode        int    `json:"statusCode"`
	SuggestedFileName string `json:"suggestedFileName"`
	Size              int64  `json:"size"`
	Downloaded        bool   `json:"downloaded"`
	DownloadedPath    string `json:"downloadedPath,omitempty"`
	FirstSeenAt       string `json:"firstSeenAt"`
	LastSeenAt        string `json:"lastSeenAt"`
}

// MonitoredRequest 表示监听页面过程中捕获到的请求
type MonitoredRequest struct {
	ID                  string            `json:"id"`
	URL                 string            `json:"url"`
	Method              string            `json:"method"`
	ResourceType        string            `json:"resourceType,omitempty"`
	MimeType            string            `json:"mimeType,omitempty"`
	StatusCode          int               `json:"statusCode"`
	Failed              bool              `json:"failed"`
	FailureText         string            `json:"failureText,omitempty"`
	RequestHeaders      map[string]string `json:"requestHeaders,omitempty"`
	ResponseHeaders     map[string]string `json:"responseHeaders,omitempty"`
	RequestBodyPreview  string            `json:"requestBodyPreview,omitempty"`
	ResponseBodyPreview string            `json:"responseBodyPreview,omitempty"`
	FirstSeenAt         string            `json:"firstSeenAt"`
}

// DownloadResourcesResult 表示批量下载结果
type DownloadResourcesResult struct {
	TaskID            string               `json:"taskId"`
	DownloadDir       string               `json:"downloadDir"`
	DownloadedIDs     []string             `json:"downloadedIds"`
	SkippedIDs        []string             `json:"skippedIds"`
	DownloadedEntries []*MonitoredResource `json:"downloadedEntries"`
}

// ResourceMonitorEvent 表示推送给前端的事件
type ResourceMonitorEvent struct {
	Type     string                   `json:"type"`
	Message  string                   `json:"message,omitempty"`
	Task     *ResourceMonitorTask     `json:"task,omitempty"`
	Resource *MonitoredResource       `json:"resource,omitempty"`
	Request  *MonitoredRequest        `json:"request,omitempty"`
	Download *DownloadResourcesResult `json:"download,omitempty"`
}
