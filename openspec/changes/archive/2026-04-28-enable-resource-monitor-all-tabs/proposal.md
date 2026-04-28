## Why

当前资源监听页面只对启动任务时绑定的单个页面对象开启 DrissionPage 监听。用户在监听过程中如果手动打开新标签页，新的请求和资源不会进入当前任务列表，导致“浏览器正在工作，但监控面板没有记录”的割裂体验。

同时，并不是所有场景都需要跨标签页监听：有些用户只想保持当前的单标签页采集，以减少噪音。因此最合适的方式不是直接替换原行为，而是提供一个明确的启动选项，并默认开启，兼顾覆盖面与可控性。

## What Changes

- 在资源监听页面启动区新增“监听所有标签页”勾选项，默认勾选。
- 将该选项纳入资源监听任务模型和启动参数，并在当前任务状态中回显。
- 扩展 Python worker 的监听逻辑：勾选时为浏览器中每个已存在和后续新增的标签页分别启动监听；未勾选时保持当前单标签页实现。

## Capabilities

### Modified Capabilities

- `resource-monitoring`: 用户启动资源监听页面任务时可以选择仅监听当前标签页，或持续监听当前浏览器会话中的全部标签页（包括后续新增标签页）。

## Impact

- 影响 `frontend/src/components/resource-monitor-tab.jsx` 的启动表单与任务状态展示。
- 影响 `app.go`、`backend/services/resource_monitor_service.go`、`backend/models/resource_monitor.go` 的任务参数与模型传递。
- 影响 `backend/services/python/resource_monitor_worker.py` 的监听器注册、暂停、恢复和停止逻辑。

## Decision Snapshot

```text
默认:
当前 tab + 后续新增 tab -> 都注册各自的 listener

关闭勾选:
仅启动任务时的原始 page listener
```
