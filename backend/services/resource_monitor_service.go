package services

import (
	"bufio"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"RequestProbe/backend/models"

	"github.com/google/uuid"
)

//go:embed python/resource_monitor_worker.py
var pythonWorkerFS embed.FS

type pythonMessage struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	OK      bool            `json:"ok"`
	Error   string          `json:"error,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Message string          `json:"message,omitempty"`
}

type pythonCommand struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type pythonWorker struct {
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	pendingMu  sync.Mutex
	pending    map[string]chan pythonMessage
	nextID     uint64
	eventFn    func(pythonMessage)
	startedAt  time.Time
	closedOnce sync.Once
}

type resourceMonitorWorker interface {
	request(ctx context.Context, cmdType string, payload interface{}, out interface{}) error
	Close()
}

func newPythonWorker(ctx context.Context, eventFn func(pythonMessage)) (*pythonWorker, error) {
	pythonExec, err := resolvePythonExecutable(ctx)
	if err != nil {
		return nil, err
	}

	scriptPath, err := ensurePythonWorkerScript()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, pythonExec, scriptPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 Python stdin 失败: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 Python stdout 失败: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 Python stderr 失败: %w", err)
	}

	worker := &pythonWorker{
		cmd:       cmd,
		stdin:     stdin,
		pending:   make(map[string]chan pythonMessage),
		eventFn:   eventFn,
		startedAt: time.Now(),
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 Python worker 失败: %w", err)
	}

	go worker.readStdout(stdout)
	go worker.readStderr(stderr)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := worker.request(pingCtx, "ping", nil, nil); err != nil {
		worker.Close()
		return nil, err
	}

	return worker, nil
}

func (w *pythonWorker) Close() {
	w.closedOnce.Do(func() {
		if w.stdin != nil {
			_ = w.stdin.Close()
		}
		if w.cmd != nil && w.cmd.Process != nil {
			_ = w.cmd.Process.Kill()
			_, _ = w.cmd.Process.Wait()
		}
	})
}

func (w *pythonWorker) readStdout(r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		var msg pythonMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		if msg.ID != "" && msg.Type == "response" {
			w.pendingMu.Lock()
			ch := w.pending[msg.ID]
			delete(w.pending, msg.ID)
			w.pendingMu.Unlock()
			if ch != nil {
				ch <- msg
			}
			continue
		}

		if w.eventFn != nil {
			w.eventFn(msg)
		}
	}
}

func (w *pythonWorker) readStderr(r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 16*1024), 1024*1024)
	for scanner.Scan() {
		if w.eventFn != nil {
			w.eventFn(pythonMessage{
				Type:    "worker_log",
				Message: scanner.Text(),
			})
		}
	}
}

func (w *pythonWorker) request(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
	id := fmt.Sprintf("cmd-%d", atomic.AddUint64(&w.nextID, 1))
	respCh := make(chan pythonMessage, 1)

	w.pendingMu.Lock()
	w.pending[id] = respCh
	w.pendingMu.Unlock()

	cmd := pythonCommand{
		ID:      id,
		Type:    cmdType,
		Payload: payload,
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("编码 Python 命令失败: %w", err)
	}
	if _, err := io.WriteString(w.stdin, string(data)+"\n"); err != nil {
		return fmt.Errorf("写入 Python 命令失败: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case msg := <-respCh:
		if !msg.OK {
			return errors.New(msg.Error)
		}
		if out != nil && len(msg.Payload) > 0 {
			if err := json.Unmarshal(msg.Payload, out); err != nil {
				return fmt.Errorf("解析 Python 响应失败: %w", err)
			}
		}
		return nil
	}
}

func resolvePythonExecutable(ctx context.Context) (string, error) {
	candidates := pythonExecutableCandidates()

	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}

		path := candidate
		if !filepath.IsAbs(candidate) {
			resolved, err := exec.LookPath(candidate)
			if err != nil {
				continue
			}
			path = resolved
		} else if _, err := os.Stat(candidate); err != nil {
			continue
		}

		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		cmd := exec.CommandContext(checkCtx, path, "-c", "import DrissionPage")
		if err := cmd.Run(); err == nil {
			cancel()
			return path, nil
		}
		cancel()
	}

	return "", errors.New("未找到可用的 Python + DrissionPage 运行环境，请先安装 DrissionPage，或设置 REQUESTPROBE_PYTHON 指向可用解释器")
}

func pythonExecutableCandidates() []string {
	candidates := []string{}
	if envPath := strings.TrimSpace(os.Getenv("REQUESTPROBE_PYTHON")); envPath != "" {
		candidates = append(candidates, envPath)
	}

	if execPath, err := os.Executable(); err == nil {
		candidates = append(candidates, bundledPythonExecutableCandidates(execPath)...)
	}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, developmentPythonExecutableCandidates(cwd)...)
	}

	candidates = append(candidates, "python3", "python")
	return uniqueNonEmptyStrings(candidates)
}

func bundledPythonExecutableCandidates(execPath string) []string {
	if strings.TrimSpace(execPath) == "" {
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		baseDir := filepath.Join(filepath.Dir(execPath), "..", "Resources", "python")
		return []string{
			filepath.Join(baseDir, "bin", "python3"),
			filepath.Join(baseDir, "bin", "python"),
			filepath.Join(baseDir, "bin", "python3.13"),
		}
	case "windows":
		baseDir := filepath.Join(filepath.Dir(execPath), "python")
		return []string{
			filepath.Join(baseDir, "python.exe"),
			filepath.Join(baseDir, "python3.exe"),
		}
	default:
		baseDir := filepath.Join(filepath.Dir(execPath), "python")
		return []string{
			filepath.Join(baseDir, "bin", "python3"),
			filepath.Join(baseDir, "bin", "python"),
			filepath.Join(baseDir, "python3"),
			filepath.Join(baseDir, "python"),
		}
	}
}

func developmentPythonExecutableCandidates(cwd string) []string {
	if strings.TrimSpace(cwd) == "" {
		return nil
	}

	envDirs := []string{
		filepath.Join(cwd, ".venv-monitor"),
		filepath.Join(cwd, ".venv"),
	}

	candidates := []string{}
	for _, envDir := range envDirs {
		candidates = append(candidates, venvExecutableCandidates(envDir)...)
	}
	return candidates
}

func venvExecutableCandidates(envDir string) []string {
	if strings.TrimSpace(envDir) == "" {
		return nil
	}

	return []string{
		filepath.Join(envDir, "bin", "python3"),
		filepath.Join(envDir, "bin", "python"),
		filepath.Join(envDir, "Scripts", "python.exe"),
		filepath.Join(envDir, "Scripts", "python3.exe"),
	}
}

func uniqueNonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func ensurePythonWorkerScript() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("获取缓存目录失败: %w", err)
	}

	workerDir := filepath.Join(cacheDir, "RequestProbe", "python-worker")
	if err := os.MkdirAll(workerDir, 0o755); err != nil {
		return "", fmt.Errorf("创建 Python worker 目录失败: %w", err)
	}

	content, err := pythonWorkerFS.ReadFile("python/resource_monitor_worker.py")
	if err != nil {
		return "", fmt.Errorf("读取内嵌 Python worker 失败: %w", err)
	}

	sum := sha256.Sum256(content)
	fileName := fmt.Sprintf("resource_monitor_worker_%s.py", hex.EncodeToString(sum[:8]))
	path := filepath.Join(workerDir, fileName)

	if err := os.WriteFile(path, content, 0o755); err != nil {
		return "", fmt.Errorf("写入 Python worker 文件失败: %w", err)
	}

	return path, nil
}

// ResourceMonitorService 提供资源监听能力
type ResourceMonitorService struct {
	mu             sync.RWMutex
	ctx            context.Context
	worker         resourceMonitorWorker
	workerFactory  func(context.Context, func(pythonMessage)) (resourceMonitorWorker, error)
	settingsPathFn func() (string, error)
	defaultRootFn  func() (string, error)
	task           *models.ResourceMonitorTask
	eventFn        func(*models.ResourceMonitorEvent)
}

// NewResourceMonitorService 创建资源监听服务
func NewResourceMonitorService() *ResourceMonitorService {
	return &ResourceMonitorService{
		workerFactory: func(ctx context.Context, eventFn func(pythonMessage)) (resourceMonitorWorker, error) {
			return newPythonWorker(ctx, eventFn)
		},
		settingsPathFn: defaultResourceMonitorSettingsPath,
		defaultRootFn:  defaultResourceMonitorSaveRootDir,
	}
}

// SetContext 设置上下文
func (s *ResourceMonitorService) SetContext(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
}

// SetEventHandler 设置事件回调
func (s *ResourceMonitorService) SetEventHandler(eventFn func(*models.ResourceMonitorEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventFn = eventFn
}

// GetCommonResourceExtensions 返回常见监听后缀
func (s *ResourceMonitorService) GetCommonResourceExtensions() []string {
	return []string{"js", "wasm", "css", "json", "map"}
}

// GetCurrentTask 返回当前任务状态
func (s *ResourceMonitorService) GetCurrentTask(ctx context.Context) *models.ResourceMonitorTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneTask(s.task)
}

// StartTask 启动任务
func (s *ResourceMonitorService) StartTask(ctx context.Context, rawURL string, extensions []string) (*models.ResourceMonitorTask, error) {
	normalizedURL, err := normalizeURL(rawURL)
	if err != nil {
		return nil, err
	}

	normalizedExts := normalizeExtensions(extensions)
	if len(normalizedExts) == 0 {
		return nil, errors.New("至少选择一个文件后缀")
	}

	saveRootDir, err := s.getEffectiveSaveRootDir()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if s.task != nil && (s.task.Status == models.ResourceMonitorStatusRunning || s.task.Status == models.ResourceMonitorStatusPaused) {
		s.mu.Unlock()
		return nil, errors.New("当前已有活跃的资源监听任务，请先结束当前任务")
	}

	if s.ctx == nil {
		s.ctx = ctx
	}

	worker, err := s.ensureWorkerLocked(ctx)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.mu.Unlock()

	taskID := uuid.NewString()
	downloadDir, err := prepareDownloadDir(saveRootDir, taskID)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"taskId":      taskID,
		"url":         normalizedURL,
		"extensions":  normalizedExts,
		"downloadDir": downloadDir,
	}

	var task models.ResourceMonitorTask
	if err := worker.request(ctx, "start_task", payload, &task); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	task.Status = models.ResourceMonitorStatusRunning
	s.task = &task
	s.emitLocked(&models.ResourceMonitorEvent{
		Type: "task_updated",
		Task: cloneTask(s.task),
	})

	return cloneTask(s.task), nil
}

// PauseTask 暂停任务
func (s *ResourceMonitorService) PauseTask(ctx context.Context) (*models.ResourceMonitorTask, error) {
	s.mu.RLock()
	worker := s.worker
	s.mu.RUnlock()

	if worker == nil {
		return nil, errors.New("当前没有资源监听任务")
	}
	var task models.ResourceMonitorTask
	if err := worker.request(ctx, "pause_task", nil, &task); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	task.Status = models.ResourceMonitorStatusPaused
	s.task = &task
	s.emitLocked(&models.ResourceMonitorEvent{Type: "task_updated", Task: cloneTask(s.task)})
	return cloneTask(s.task), nil
}

// ResumeTask 恢复任务
func (s *ResourceMonitorService) ResumeTask(ctx context.Context) (*models.ResourceMonitorTask, error) {
	s.mu.RLock()
	worker := s.worker
	s.mu.RUnlock()

	if worker == nil {
		return nil, errors.New("当前没有资源监听任务")
	}
	var task models.ResourceMonitorTask
	if err := worker.request(ctx, "resume_task", nil, &task); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	task.Status = models.ResourceMonitorStatusRunning
	s.task = &task
	s.emitLocked(&models.ResourceMonitorEvent{Type: "task_updated", Task: cloneTask(s.task)})
	return cloneTask(s.task), nil
}

// EndTask 结束任务
func (s *ResourceMonitorService) EndTask(ctx context.Context) (*models.ResourceMonitorTask, error) {
	s.mu.RLock()
	worker := s.worker
	s.mu.RUnlock()

	if worker == nil {
		return nil, errors.New("当前没有资源监听任务")
	}
	var task models.ResourceMonitorTask
	if err := worker.request(ctx, "end_task", nil, &task); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	task.Status = models.ResourceMonitorStatusEnded
	s.task = &task
	s.emitLocked(&models.ResourceMonitorEvent{Type: "task_updated", Task: cloneTask(s.task)})
	return cloneTask(s.task), nil
}

// DownloadResources 下载资源
func (s *ResourceMonitorService) DownloadResources(ctx context.Context, resourceIDs []string) (*models.DownloadResourcesResult, error) {
	s.mu.RLock()
	worker := s.worker
	s.mu.RUnlock()

	if worker == nil {
		return nil, errors.New("当前没有资源监听任务")
	}

	ids := normalizeResourceIDs(resourceIDs)
	if len(ids) == 0 {
		return nil, errors.New("未选择任何资源")
	}

	var result models.DownloadResourcesResult
	if err := worker.request(ctx, "download_resources", map[string]interface{}{"resourceIds": ids}, &result); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.task != nil {
		for _, downloaded := range result.DownloadedEntries {
			for _, item := range s.task.Resources {
				if item.ID == downloaded.ID {
					item.Downloaded = downloaded.Downloaded
					item.DownloadedPath = downloaded.DownloadedPath
					item.LastSeenAt = downloaded.LastSeenAt
				}
			}
		}
		s.task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}

	s.emitLocked(&models.ResourceMonitorEvent{
		Type:     "resources_downloaded",
		Task:     cloneTask(s.task),
		Download: cloneDownloadResult(&result),
	})

	return cloneDownloadResult(&result), nil
}

// OpenDownloadDir 用 Finder 或 VS Code 打开目录
func (s *ResourceMonitorService) OpenDownloadDir(ctx context.Context, opener string) error {
	s.mu.RLock()
	task := cloneTask(s.task)
	s.mu.RUnlock()

	if task == nil || task.DownloadDir == "" {
		return errors.New("当前没有可打开的下载目录")
	}

	switch strings.ToLower(strings.TrimSpace(opener)) {
	case "finder":
		return exec.CommandContext(ctx, "open", task.DownloadDir).Run()
	case "vscode":
		codePath, err := exec.LookPath("code")
		if err != nil {
			return errors.New("未找到 VS Code CLI，请先在 VS Code 中安装 `code` 命令")
		}
		return exec.CommandContext(ctx, codePath, task.DownloadDir).Run()
	default:
		return errors.New("不支持的打开方式")
	}
}

func (s *ResourceMonitorService) ensureWorkerLocked(ctx context.Context) (resourceMonitorWorker, error) {
	if s.worker != nil {
		return s.worker, nil
	}
	worker, err := s.workerFactory(ctx, s.handleWorkerEvent)
	if err != nil {
		return nil, err
	}
	s.worker = worker
	return worker, nil
}

func (s *ResourceMonitorService) handleWorkerEvent(msg pythonMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch msg.Type {
	case "task_updated":
		var task models.ResourceMonitorTask
		if err := json.Unmarshal(msg.Payload, &task); err == nil {
			s.task = &task
			s.sortResourcesLocked()
			s.sortRequestsLocked()
			s.emitLocked(&models.ResourceMonitorEvent{
				Type: msg.Type,
				Task: cloneTask(s.task),
			})
		}
	case "resource_detected":
		var payload struct {
			Task     *models.ResourceMonitorTask `json:"task"`
			Resource *models.MonitoredResource   `json:"resource"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if payload.Task != nil {
			s.task = payload.Task
		}
		if s.task != nil && payload.Resource != nil {
			replaced := false
			for i, item := range s.task.Resources {
				if item.ID == payload.Resource.ID {
					s.task.Resources[i] = payload.Resource
					replaced = true
					break
				}
			}
			if !replaced {
				s.task.Resources = append(s.task.Resources, payload.Resource)
			}
			s.task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			s.sortResourcesLocked()
		}
		s.emitLocked(&models.ResourceMonitorEvent{
			Type:     "resource_detected",
			Task:     cloneTask(s.task),
			Resource: cloneResource(payload.Resource),
		})
	case "request_detected":
		var payload struct {
			Task    *models.ResourceMonitorTask `json:"task"`
			Request *models.MonitoredRequest    `json:"request"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if payload.Task != nil {
			s.task = payload.Task
		}
		if s.task != nil && payload.Request != nil {
			replaced := false
			for i, item := range s.task.Requests {
				if item.ID == payload.Request.ID {
					s.task.Requests[i] = payload.Request
					replaced = true
					break
				}
			}
			if !replaced {
				s.task.Requests = append(s.task.Requests, payload.Request)
			}
			s.task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			s.sortRequestsLocked()
		}
		s.emitLocked(&models.ResourceMonitorEvent{
			Type:    "request_detected",
			Task:    cloneTask(s.task),
			Request: cloneRequest(payload.Request),
		})
	case "resources_downloaded":
		var result models.DownloadResourcesResult
		if err := json.Unmarshal(msg.Payload, &result); err != nil {
			return
		}
		if s.task != nil {
			for _, downloaded := range result.DownloadedEntries {
				for _, item := range s.task.Resources {
					if item.ID == downloaded.ID {
						item.Downloaded = downloaded.Downloaded
						item.DownloadedPath = downloaded.DownloadedPath
					}
				}
			}
		}
		s.emitLocked(&models.ResourceMonitorEvent{
			Type:     "resources_downloaded",
			Task:     cloneTask(s.task),
			Download: cloneDownloadResult(&result),
		})
	case "worker_log":
		if msg.Message != "" {
			s.emitLocked(&models.ResourceMonitorEvent{
				Type:    "worker_log",
				Message: msg.Message,
				Task:    cloneTask(s.task),
			})
		}
	}
}

func (s *ResourceMonitorService) sortResourcesLocked() {
	if s.task == nil {
		return
	}
	sort.SliceStable(s.task.Resources, func(i, j int) bool {
		return s.task.Resources[i].FirstSeenAt > s.task.Resources[j].FirstSeenAt
	})
}

func (s *ResourceMonitorService) sortRequestsLocked() {
	if s.task == nil {
		return
	}
	sort.SliceStable(s.task.Requests, func(i, j int) bool {
		return s.task.Requests[i].FirstSeenAt > s.task.Requests[j].FirstSeenAt
	})
}

func (s *ResourceMonitorService) emitLocked(event *models.ResourceMonitorEvent) {
	if event == nil || s.eventFn == nil {
		return
	}
	s.eventFn(event)
}

func normalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", errors.New("URL 格式无效")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("URL 必须以 http:// 或 https:// 开头")
	}
	if parsed.Host == "" {
		return "", errors.New("URL 缺少主机名")
	}
	return parsed.String(), nil
}

func normalizeExtensions(extensions []string) []string {
	set := make(map[string]struct{})
	for _, item := range extensions {
		ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(item, ".")))
		if ext == "" {
			continue
		}
		set[ext] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for ext := range set {
		result = append(result, ext)
	}
	sort.Strings(result)
	return result
}

func normalizeResourceIDs(ids []string) []string {
	set := make(map[string]struct{})
	for _, id := range ids {
		normalized := strings.TrimSpace(id)
		if normalized == "" {
			continue
		}
		set[normalized] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for id := range set {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func prepareDownloadDir(rootDir, taskID string) (string, error) {
	normalizedRoot, err := normalizeSaveRootDir(rootDir)
	if err != nil {
		return "", err
	}
	downloadDir := filepath.Join(normalizedRoot, taskID)
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		return "", fmt.Errorf("创建下载目录失败: %w", err)
	}
	return downloadDir, nil
}

func cloneTask(task *models.ResourceMonitorTask) *models.ResourceMonitorTask {
	if task == nil {
		return nil
	}
	cloned := *task
	cloned.SelectedExtensions = append([]string(nil), task.SelectedExtensions...)
	cloned.Resources = make([]*models.MonitoredResource, 0, len(task.Resources))
	for _, item := range task.Resources {
		cloned.Resources = append(cloned.Resources, cloneResource(item))
	}
	cloned.Requests = make([]*models.MonitoredRequest, 0, len(task.Requests))
	for _, item := range task.Requests {
		cloned.Requests = append(cloned.Requests, cloneRequest(item))
	}
	return &cloned
}

func cloneResource(resource *models.MonitoredResource) *models.MonitoredResource {
	if resource == nil {
		return nil
	}
	cloned := *resource
	return &cloned
}

func cloneRequest(request *models.MonitoredRequest) *models.MonitoredRequest {
	if request == nil {
		return nil
	}
	cloned := *request
	if request.RequestHeaders != nil {
		cloned.RequestHeaders = make(map[string]string, len(request.RequestHeaders))
		for key, value := range request.RequestHeaders {
			cloned.RequestHeaders[key] = value
		}
	}
	if request.ResponseHeaders != nil {
		cloned.ResponseHeaders = make(map[string]string, len(request.ResponseHeaders))
		for key, value := range request.ResponseHeaders {
			cloned.ResponseHeaders[key] = value
		}
	}
	return &cloned
}

func cloneDownloadResult(result *models.DownloadResourcesResult) *models.DownloadResourcesResult {
	if result == nil {
		return nil
	}
	cloned := *result
	cloned.DownloadedIDs = append([]string(nil), result.DownloadedIDs...)
	cloned.SkippedIDs = append([]string(nil), result.SkippedIDs...)
	cloned.DownloadedEntries = make([]*models.MonitoredResource, 0, len(result.DownloadedEntries))
	for _, item := range result.DownloadedEntries {
		cloned.DownloadedEntries = append(cloned.DownloadedEntries, cloneResource(item))
	}
	return &cloned
}
