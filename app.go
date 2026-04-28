package main

import (
	"context"
	"fmt"

	"RequestProbe/backend/models"
	"RequestProbe/backend/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                    context.Context
	requestService         *services.RequestService
	resourceMonitorService *services.ResourceMonitorService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		requestService:         services.NewRequestService(),
		resourceMonitorService: services.NewResourceMonitorService(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.resourceMonitorService.SetContext(ctx)
	a.resourceMonitorService.SetEventHandler(func(event *models.ResourceMonitorEvent) {
		runtime.EventsEmit(a.ctx, "resource-monitor-event", event)
	})
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s from RequestProbe!", name)
}

// ParseRequest 解析请求
func (a *App) ParseRequest(input string) (*models.ParsedRequest, error) {
	return a.requestService.ParseRequest(a.ctx, input)
}

// ParseRequestWithType 使用指定类型解析请求
func (a *App) ParseRequestWithType(input, inputType string) (*models.ParsedRequest, error) {
	return a.requestService.ParseRequestWithType(a.ctx, input, inputType)
}

// DetectInputType 检测输入类型
func (a *App) DetectInputType(input string) string {
	return a.requestService.DetectInputType(a.ctx, input)
}

// GeneratePythonCode 生成Python代码
func (a *App) GeneratePythonCode(request *models.ParsedRequest) string {
	return a.requestService.GeneratePythonCode(a.ctx, request)
}

// TestSingleRequest 测试单个请求
func (a *App) TestSingleRequest(request *models.ParsedRequest, config *models.ValidationConfig) (*models.ResponseData, error) {
	return a.requestService.TestSingleRequest(a.ctx, request, config)
}

// TestFieldNecessity 测试字段必要性
func (a *App) TestFieldNecessity(request *models.ParsedRequest, config *models.ValidationConfig) (*models.BatchTestResult, error) {
	// 创建进度回调函数，同时发送到前端和控制台
	progressCallback := func(progress *models.TestProgress) {
		// 打印进度信息到控制台
		fmt.Printf("Progress: %s (%.1f%%)\n", progress.Message, progress.Progress)
		// 发送进度事件到前端
		runtime.EventsEmit(a.ctx, "test-progress", progress)
	}

	return a.requestService.TestFieldNecessity(a.ctx, request, config, progressCallback)
}

// TestFieldNecessityWithProgress 测试字段必要性（带前端进度回调）
func (a *App) TestFieldNecessityWithProgress(request *models.ParsedRequest, config *models.ValidationConfig) (*models.BatchTestResult, error) {
	// 创建进度回调函数，通过 Wails 事件系统发送到前端
	progressCallback := func(progress *models.TestProgress) {
		// 发送进度事件到前端
		runtime.EventsEmit(a.ctx, "test-progress", progress)
	}

	return a.requestService.TestFieldNecessity(a.ctx, request, config, progressCallback)
}

// ValidateExpression 验证表达式
func (a *App) ValidateExpression(expression string) error {
	return a.requestService.ValidateExpression(a.ctx, expression)
}

// GetExpressionTemplates 获取表达式模板
func (a *App) GetExpressionTemplates() []models.ExpressionTemplate {
	return a.requestService.GetExpressionTemplates(a.ctx)
}

// GetExpressionTemplatesByCategory 按分类获取模板
func (a *App) GetExpressionTemplatesByCategory(category string) []models.ExpressionTemplate {
	return a.requestService.GetExpressionTemplatesByCategory(a.ctx, category)
}

// GetExpressionCategories 获取表达式分类
func (a *App) GetExpressionCategories() []string {
	return a.requestService.GetExpressionCategories(a.ctx)
}

// AddExpressionTemplate 添加表达式模板
func (a *App) AddExpressionTemplate(template models.ExpressionTemplate) error {
	return a.requestService.AddExpressionTemplate(a.ctx, template)
}

// UpdateExpressionTemplate 更新表达式模板
func (a *App) UpdateExpressionTemplate(template models.ExpressionTemplate) error {
	return a.requestService.UpdateExpressionTemplate(a.ctx, template)
}

// DeleteExpressionTemplate 删除表达式模板
func (a *App) DeleteExpressionTemplate(id string) error {
	return a.requestService.DeleteExpressionTemplate(a.ctx, id)
}

// GetDefaultValidationConfig 获取默认验证配置
func (a *App) GetDefaultValidationConfig() *models.ValidationConfig {
	return a.requestService.GetDefaultValidationConfig(a.ctx)
}

// GetRequestSummary 获取请求摘要信息
func (a *App) GetRequestSummary(request *models.ParsedRequest) map[string]interface{} {
	return a.requestService.GetRequestSummary(a.ctx, request)
}

// GetTestStatistics 获取测试统计信息
func (a *App) GetTestStatistics(result *models.BatchTestResult) map[string]interface{} {
	return a.requestService.GetTestStatistics(a.ctx, result)
}

// TestRequestOnly 仅测试请求（不进行字段必要性分析）
func (a *App) TestRequestOnly(request *models.ParsedRequest, config *models.ValidationConfig) (*models.ResponseData, error) {
	return a.requestService.TestSingleRequest(a.ctx, request, config)
}

// GetSupportedEncodings 获取支持的编码列表
func (a *App) GetSupportedEncodings() []string {
	return a.requestService.GetSupportedEncodings(a.ctx)
}

// GetCommonEncodings 获取常用编码列表
func (a *App) GetCommonEncodings() []string {
	return a.requestService.GetCommonEncodings(a.ctx)
}

// DetectEncoding 检测响应编码
func (a *App) DetectEncoding(responseBody string, calibrationText string) (string, error) {
	return a.requestService.DetectEncoding(a.ctx, []byte(responseBody), calibrationText)
}

// DecodeResponse 使用指定编码解码响应
func (a *App) DecodeResponse(responseBody string, encodingName string) (string, error) {
	return a.requestService.DecodeResponse(a.ctx, []byte(responseBody), encodingName)
}

// DetectEncodingFromResponse 从响应数据中检测编码（使用原始字节数据）
func (a *App) DetectEncodingFromResponse(response *models.ResponseData, calibrationText string) (string, error) {
	return a.requestService.DetectEncodingFromResponse(a.ctx, response, calibrationText)
}

// DecodeResponseFromResponse 从响应数据中解码（使用原始字节数据）
func (a *App) DecodeResponseFromResponse(response *models.ResponseData, encodingName string) (string, error) {
	return a.requestService.DecodeResponseFromResponse(a.ctx, response, encodingName)
}

// AutoDetectEncodingFromResponse 自动检测响应编码并转换
func (a *App) AutoDetectEncodingFromResponse(response *models.ResponseData) (map[string]interface{}, error) {
	decodedText, detectedEncoding, err := a.requestService.AutoDetectEncodingFromResponse(a.ctx, response)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"decodedText":      decodedText,
		"detectedEncoding": detectedEncoding,
	}, nil
}

// ParseSimpleRequest 简单的请求解析测试（保持向后兼容）
func (a *App) ParseSimpleRequest(input string) map[string]interface{} {
	// 尝试解析真实请求
	request, err := a.requestService.ParseRequest(a.ctx, input)
	if err != nil {
		return map[string]interface{}{
			"input":  input,
			"error":  err.Error(),
			"status": "failed",
		}
	}

	return map[string]interface{}{
		"input":       input,
		"method":      request.Method,
		"url":         request.URL,
		"headers":     request.Headers,
		"cookies":     request.Cookies,
		"body":        request.Body,
		"contentType": request.ContentType,
		"status":      "success",
	}
}

// GetCommonResourceExtensions 获取常见资源后缀
func (a *App) GetCommonResourceExtensions() []string {
	return a.resourceMonitorService.GetCommonResourceExtensions()
}

// GetResourceMonitorTask 获取当前资源监听任务
func (a *App) GetResourceMonitorTask() *models.ResourceMonitorTask {
	return a.resourceMonitorService.GetCurrentTask(a.ctx)
}

// StartResourceMonitor 启动资源监听任务
func (a *App) StartResourceMonitor(rawURL string, extensions []string, listenAllTabs bool) (*models.ResourceMonitorTask, error) {
	return a.resourceMonitorService.StartTask(a.ctx, rawURL, extensions, listenAllTabs)
}

// PauseResourceMonitor 暂停资源监听
func (a *App) PauseResourceMonitor() (*models.ResourceMonitorTask, error) {
	return a.resourceMonitorService.PauseTask(a.ctx)
}

// ResumeResourceMonitor 恢复资源监听
func (a *App) ResumeResourceMonitor() (*models.ResourceMonitorTask, error) {
	return a.resourceMonitorService.ResumeTask(a.ctx)
}

// EndResourceMonitor 结束资源监听任务
func (a *App) EndResourceMonitor() (*models.ResourceMonitorTask, error) {
	return a.resourceMonitorService.EndTask(a.ctx)
}

// DownloadSelectedResources 下载选中的资源
func (a *App) DownloadSelectedResources(resourceIDs []string) (*models.DownloadResourcesResult, error) {
	return a.resourceMonitorService.DownloadResources(a.ctx, resourceIDs)
}

// DownloadSelectedRequests 下载选中的请求记录
func (a *App) DownloadSelectedRequests(requestIDs []string) (*models.DownloadRequestsResult, error) {
	return a.resourceMonitorService.DownloadRequests(a.ctx, requestIDs)
}

// OpenResourceMonitorDownloadDir 打开资源监听下载目录
func (a *App) OpenResourceMonitorDownloadDir(opener string) error {
	return a.resourceMonitorService.OpenDownloadDir(a.ctx, opener)
}

// GetResourceMonitorSettings 获取资源监听设置
func (a *App) GetResourceMonitorSettings() (*models.ResourceMonitorSettings, error) {
	return a.resourceMonitorService.GetSettings(a.ctx)
}

// UpdateResourceMonitorSaveRoot 更新资源监听默认保存目录
func (a *App) UpdateResourceMonitorSaveRoot(saveRoot string) (*models.ResourceMonitorSettings, error) {
	return a.resourceMonitorService.UpdateSaveRoot(a.ctx, saveRoot)
}

// ResetResourceMonitorSaveRoot 重置资源监听默认保存目录
func (a *App) ResetResourceMonitorSaveRoot() (*models.ResourceMonitorSettings, error) {
	return a.resourceMonitorService.ResetSaveRoot(a.ctx)
}

// ChooseResourceMonitorSaveRoot 打开系统目录选择器
func (a *App) ChooseResourceMonitorSaveRoot() (string, error) {
	settings, err := a.resourceMonitorService.GetSettings(a.ctx)
	if err != nil {
		return "", err
	}

	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "选择资源监听页面保存目录",
		DefaultDirectory:     settings.SaveRootDir,
		CanCreateDirectories: true,
		ShowHiddenFiles:      true,
	})
}
