package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"RequestProbe/backend/models"
)

type fakeMonitorWorker struct {
	requestFn func(ctx context.Context, cmdType string, payload interface{}, out interface{}) error
}

func (f *fakeMonitorWorker) request(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
	return f.requestFn(ctx, cmdType, payload, out)
}

func (f *fakeMonitorWorker) Close() {}

func TestStartTaskReturnsWhenEventArrivesBeforeResponse(t *testing.T) {
	svc := NewResourceMonitorService()
	svc.workerFactory = func(ctx context.Context, eventFn func(pythonMessage)) (resourceMonitorWorker, error) {
		return &fakeMonitorWorker{
			requestFn: func(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
				eventFn(pythonMessage{
					Type:    "task_updated",
					Payload: mustMarshalRaw(t, makeTask(models.ResourceMonitorStatusRunning)),
				})

				task := out.(*models.ResourceMonitorTask)
				*task = *makeTask(models.ResourceMonitorStatusRunning)
				return nil
			},
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})
	var (
		task *models.ResourceMonitorTask
		err  error
	)

	go func() {
		task, err = svc.StartTask(ctx, "https://example.com", []string{"js"}, true)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("StartTask 被 task_updated 事件阻塞，没有及时返回")
	}

	if err != nil {
		t.Fatalf("StartTask 返回错误: %v", err)
	}
	if task == nil || task.Status != models.ResourceMonitorStatusRunning {
		t.Fatalf("StartTask 返回任务异常: %#v", task)
	}
}

func TestPauseTaskReturnsWhenEventArrivesBeforeResponse(t *testing.T) {
	svc := NewResourceMonitorService()
	svc.task = makeTask(models.ResourceMonitorStatusRunning)
	svc.worker = &fakeMonitorWorker{
		requestFn: func(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
			svc.handleWorkerEvent(pythonMessage{
				Type:    "task_updated",
				Payload: mustMarshalRaw(t, makeTask(models.ResourceMonitorStatusPaused)),
			})

			task := out.(*models.ResourceMonitorTask)
			*task = *makeTask(models.ResourceMonitorStatusPaused)
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})
	var err error
	go func() {
		_, err = svc.PauseTask(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("PauseTask 被 task_updated 事件阻塞，没有及时返回")
	}

	if err != nil {
		t.Fatalf("PauseTask 返回错误: %v", err)
	}
}

func TestNormalizeURLEmptyIsAllowed(t *testing.T) {
	got, err := normalizeURL("")
	if err != nil {
		t.Fatalf("空 URL 不应报错: %v", err)
	}
	if got != "" {
		t.Fatalf("空 URL 应保持为空，实际为 %q", got)
	}
}

func TestStartTaskAllowsEmptyURL(t *testing.T) {
	svc := NewResourceMonitorService()
	svc.workerFactory = func(ctx context.Context, eventFn func(pythonMessage)) (resourceMonitorWorker, error) {
		return &fakeMonitorWorker{
			requestFn: func(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
				data, ok := payload.(map[string]interface{})
				if !ok {
					t.Fatalf("payload 类型异常: %#v", payload)
				}
				if data["url"] != "" {
					t.Fatalf("空 URL 启动时应传空字符串，实际为 %#v", data["url"])
				}
				task := out.(*models.ResourceMonitorTask)
				*task = *makeTask(models.ResourceMonitorStatusRunning)
				task.URL = ""
				return nil
			},
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	task, err := svc.StartTask(ctx, "", []string{"js"}, true)
	if err != nil {
		t.Fatalf("空 URL 启动不应失败: %v", err)
	}
	if task == nil {
		t.Fatal("空 URL 启动应返回任务")
	}
	if task.URL != "" {
		t.Fatalf("任务 URL 应为空，实际为 %q", task.URL)
	}
}

func TestHandleWorkerEventTracksRequests(t *testing.T) {
	svc := NewResourceMonitorService()
	svc.task = makeTask(models.ResourceMonitorStatusRunning)

	request := &models.MonitoredRequest{
		ID:           "req-1",
		URL:          "https://example.com/api/users",
		Method:       "POST",
		ResourceType: "XHR",
		StatusCode:   201,
		FirstSeenAt:  time.Now().UTC().Format(time.RFC3339Nano),
	}

	svc.handleWorkerEvent(pythonMessage{
		Type: "request_detected",
		Payload: mustMarshalRaw(t, map[string]interface{}{
			"task":    svc.task,
			"request": request,
		}),
	})

	if svc.task == nil {
		t.Fatal("任务不应为空")
	}
	if len(svc.task.Requests) != 1 {
		t.Fatalf("应记录 1 个请求，实际为 %d", len(svc.task.Requests))
	}
	if svc.task.Requests[0].ID != request.ID {
		t.Fatalf("请求 ID 不匹配: %#v", svc.task.Requests[0])
	}
}

func TestDownloadRequestsUpdatesTaskState(t *testing.T) {
	svc := NewResourceMonitorService()
	svc.task = makeTask(models.ResourceMonitorStatusRunning)
	svc.task.Requests = []*models.MonitoredRequest{
		{
			ID:                "req-1",
			URL:               "https://example.com/api/users",
			Method:            "GET",
			SuggestedFileName: "get-users-req-1.json",
			FirstSeenAt:       time.Now().UTC().Format(time.RFC3339Nano),
		},
	}
	svc.worker = &fakeMonitorWorker{
		requestFn: func(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
			if cmdType != "download_requests" {
				t.Fatalf("命令类型错误: %s", cmdType)
			}
			data, ok := payload.(map[string]interface{})
			if !ok {
				t.Fatalf("payload 类型异常: %#v", payload)
			}
			ids, ok := data["requestIds"].([]string)
			if !ok || len(ids) != 1 || ids[0] != "req-1" {
				t.Fatalf("requestIds 传递错误: %#v", data["requestIds"])
			}

			result := out.(*models.DownloadRequestsResult)
			*result = models.DownloadRequestsResult{
				TaskID:        svc.task.TaskID,
				DownloadDir:   filepath.Join(svc.task.DownloadDir, "requests"),
				DownloadedIDs: []string{"req-1"},
				DownloadedEntries: []*models.MonitoredRequest{
					{
						ID:             "req-1",
						Downloaded:     true,
						DownloadedPath: filepath.Join(svc.task.DownloadDir, "requests", "get-users-req-1.json"),
					},
				},
			}
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := svc.DownloadRequests(ctx, []string{"req-1"})
	if err != nil {
		t.Fatalf("DownloadRequests 返回错误: %v", err)
	}
	if result == nil || len(result.DownloadedIDs) != 1 {
		t.Fatalf("下载结果异常: %#v", result)
	}
	if !svc.task.Requests[0].Downloaded {
		t.Fatal("任务中的请求下载状态未更新")
	}
	if svc.task.Requests[0].DownloadedPath == "" {
		t.Fatal("任务中的请求下载路径未更新")
	}
}

func TestUpdateSaveRootPersistsAndStartTaskUsesConfiguredRoot(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.json")
	customRoot := filepath.Join(tempDir, "custom-downloads")

	svc := NewResourceMonitorService()
	svc.settingsPathFn = func() (string, error) { return settingsPath, nil }
	svc.defaultRootFn = func() (string, error) { return filepath.Join(tempDir, "default-downloads"), nil }
	svc.workerFactory = func(ctx context.Context, eventFn func(pythonMessage)) (resourceMonitorWorker, error) {
		return &fakeMonitorWorker{
			requestFn: func(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
				data, ok := payload.(map[string]interface{})
				if !ok {
					t.Fatalf("payload 类型异常: %#v", payload)
				}
				downloadDir, _ := data["downloadDir"].(string)
				if filepath.Dir(downloadDir) != customRoot {
					t.Fatalf("downloadDir 应位于自定义根目录下，实际为 %q", downloadDir)
				}

				task := out.(*models.ResourceMonitorTask)
				*task = *makeTask(models.ResourceMonitorStatusRunning)
				task.DownloadDir = downloadDir
				return nil
			},
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	settings, err := svc.UpdateSaveRoot(ctx, customRoot)
	if err != nil {
		t.Fatalf("UpdateSaveRoot 返回错误: %v", err)
	}
	if settings.SaveRootDir != customRoot {
		t.Fatalf("保存目录未更新，实际为 %q", settings.SaveRootDir)
	}

	task, err := svc.StartTask(ctx, "https://example.com", []string{"js"}, true)
	if err != nil {
		t.Fatalf("StartTask 返回错误: %v", err)
	}
	if task == nil || filepath.Dir(task.DownloadDir) != customRoot {
		t.Fatalf("任务下载目录未使用自定义根目录: %#v", task)
	}
}

func TestResetSaveRootFallsBackToDefault(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.json")
	defaultRoot := filepath.Join(tempDir, "default-downloads")

	svc := NewResourceMonitorService()
	svc.settingsPathFn = func() (string, error) { return settingsPath, nil }
	svc.defaultRootFn = func() (string, error) { return defaultRoot, nil }

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := svc.UpdateSaveRoot(ctx, filepath.Join(tempDir, "custom-downloads")); err != nil {
		t.Fatalf("预设自定义目录失败: %v", err)
	}

	settings, err := svc.ResetSaveRoot(ctx)
	if err != nil {
		t.Fatalf("ResetSaveRoot 返回错误: %v", err)
	}
	if settings.SaveRootDir != defaultRoot {
		t.Fatalf("重置后应回退到默认目录，实际为 %q", settings.SaveRootDir)
	}
}

func TestStartTaskDefaultsToConfiguredListenScope(t *testing.T) {
	svc := NewResourceMonitorService()
	svc.workerFactory = func(ctx context.Context, eventFn func(pythonMessage)) (resourceMonitorWorker, error) {
		return &fakeMonitorWorker{
			requestFn: func(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
				data, ok := payload.(map[string]interface{})
				if !ok {
					t.Fatalf("payload 类型异常: %#v", payload)
				}

				listenAllTabs, ok := data["listenAllTabs"].(bool)
				if !ok {
					t.Fatalf("listenAllTabs 类型异常: %#v", data["listenAllTabs"])
				}
				if !listenAllTabs {
					t.Fatal("默认启动应传递 listenAllTabs=true")
				}

				task := out.(*models.ResourceMonitorTask)
				*task = *makeTask(models.ResourceMonitorStatusRunning)
				task.ListenAllTabs = listenAllTabs
				return nil
			},
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	task, err := svc.StartTask(ctx, "https://example.com", []string{"js"}, true)
	if err != nil {
		t.Fatalf("StartTask 返回错误: %v", err)
	}
	if task == nil || !task.ListenAllTabs {
		t.Fatalf("任务应记录 listenAllTabs=true，实际为 %#v", task)
	}
}

func TestStartTaskCanDisableAllTabsListening(t *testing.T) {
	svc := NewResourceMonitorService()
	svc.workerFactory = func(ctx context.Context, eventFn func(pythonMessage)) (resourceMonitorWorker, error) {
		return &fakeMonitorWorker{
			requestFn: func(ctx context.Context, cmdType string, payload interface{}, out interface{}) error {
				data, ok := payload.(map[string]interface{})
				if !ok {
					t.Fatalf("payload 类型异常: %#v", payload)
				}

				listenAllTabs, ok := data["listenAllTabs"].(bool)
				if !ok {
					t.Fatalf("listenAllTabs 类型异常: %#v", data["listenAllTabs"])
				}
				if listenAllTabs {
					t.Fatal("关闭全标签页监听时应传递 listenAllTabs=false")
				}

				task := out.(*models.ResourceMonitorTask)
				*task = *makeTask(models.ResourceMonitorStatusRunning)
				task.ListenAllTabs = listenAllTabs
				return nil
			},
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	task, err := svc.StartTask(ctx, "https://example.com", []string{"js"}, false)
	if err != nil {
		t.Fatalf("StartTask 返回错误: %v", err)
	}
	if task == nil || task.ListenAllTabs {
		t.Fatalf("任务应记录 listenAllTabs=false，实际为 %#v", task)
	}
}

func TestDevelopmentPythonExecutableCandidatesPreferMonitorEnv(t *testing.T) {
	cwd := filepath.Join(string(filepath.Separator), "tmp", "repo")
	candidates := developmentPythonExecutableCandidates(cwd)

	if len(candidates) < 6 {
		t.Fatalf("候选列表过短: %#v", candidates)
	}
	if !strings.Contains(candidates[0], filepath.Join(".venv-monitor", "bin", "python3")) {
		t.Fatalf("应优先返回 .venv-monitor/bin/python3，实际为 %q", candidates[0])
	}
	if !strings.Contains(candidates[4], filepath.Join(".venv", "bin", "python3")) {
		t.Fatalf("应在 .venv-monitor 之后回退到 .venv，实际为 %#v", candidates)
	}
}

func TestPythonExecutableCandidatesPrioritizeEnvVar(t *testing.T) {
	original, hadOriginal := os.LookupEnv("REQUESTPROBE_PYTHON")
	if err := os.Setenv("REQUESTPROBE_PYTHON", "/custom/python"); err != nil {
		t.Fatalf("设置环境变量失败: %v", err)
	}
	defer func() {
		if hadOriginal {
			_ = os.Setenv("REQUESTPROBE_PYTHON", original)
			return
		}
		_ = os.Unsetenv("REQUESTPROBE_PYTHON")
	}()

	candidates := pythonExecutableCandidates()
	if len(candidates) == 0 {
		t.Fatal("应返回至少一个候选解释器")
	}
	if candidates[0] != "/custom/python" {
		t.Fatalf("环境变量应排在第一位，实际候选为 %#v", candidates)
	}
}

func TestUniqueNonEmptyStringsPreservesFirstOccurrence(t *testing.T) {
	got := uniqueNonEmptyStrings([]string{"", "python3", "python3", " python ", "python"})
	want := []string{"python3", "python"}
	if len(got) != len(want) {
		t.Fatalf("去重结果长度错误: got=%#v want=%#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("去重结果错误: got=%#v want=%#v", got, want)
		}
	}
}

func makeTask(status models.ResourceMonitorStatus) *models.ResourceMonitorTask {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return &models.ResourceMonitorTask{
		TaskID:             "task-1",
		URL:                "https://example.com",
		Status:             status,
		SelectedExtensions: []string{"js"},
		ListenAllTabs:      true,
		DownloadDir:        "/tmp/task-1",
		CreatedAt:          now,
		UpdatedAt:          now,
		Resources:          []*models.MonitoredResource{},
		Requests:           []*models.MonitoredRequest{},
	}
}

func mustMarshalRaw(t *testing.T, value interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	return data
}
