## Why

当前资源监听页面已经支持对命中的资源进行勾选、全选和批量下载，但请求列表只能查看，不能导出。用户在分析接口行为、失败请求、请求头和请求体时，仍然需要手工复制内容，无法像资源一样直接沉淀到本地目录。

这会让“资源监听”和“请求监听”在能力上出现明显断层。既然请求记录本身已经由同一个任务持续采集，就应该允许用户在同一界面中对请求包执行批量下载。

## What Changes

- 在请求列表中新增勾选、全选和批量下载操作。
- 为每个请求记录补充建议文件名、下载状态和下载路径字段。
- 扩展 Python worker 和 Go 服务，支持将选中的请求记录以 JSON 文件批量落盘到当前任务目录下的 `requests/` 子目录。

## Capabilities

### Modified Capabilities

- `resource-monitoring`: 用户现在不仅可以批量下载资源，也可以批量下载请求包。

## Impact

- 影响 `frontend/src/components/resource-monitor-tab.jsx` 的请求表格交互。
- 影响 `backend/models/resource_monitor.go`、`backend/services/resource_monitor_service.go`、`app.go` 的请求下载接口和模型。
- 影响 `backend/services/python/resource_monitor_worker.py` 的请求记录存储和请求包落盘逻辑。
