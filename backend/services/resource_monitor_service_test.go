package services

import (
	"context"
	"encoding/json"
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
		task, err = svc.StartTask(ctx, "https://example.com", []string{"js"})
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

	task, err := svc.StartTask(ctx, "", []string{"js"})
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

func makeTask(status models.ResourceMonitorStatus) *models.ResourceMonitorTask {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return &models.ResourceMonitorTask{
		TaskID:             "task-1",
		URL:                "https://example.com",
		Status:             status,
		SelectedExtensions: []string{"js"},
		DownloadDir:        "/tmp/task-1",
		CreatedAt:          now,
		UpdatedAt:          now,
		Resources:          []*models.MonitoredResource{},
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
