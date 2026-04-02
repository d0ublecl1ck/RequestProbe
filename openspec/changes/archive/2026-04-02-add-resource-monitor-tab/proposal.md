## Why

当前应用只覆盖请求解析、单次验证和字段必要性分析，缺少“在真实浏览器环境下观察前端资源加载并选择性落盘”的工作流。为了让用户在桌面端直接完成资源监听、筛选、下载和目录打开，需要新增一个以 DrissionPage 为核心的浏览器监听能力，并通过 Go 调用 Python 持久化管理浏览器任务。

## What Changes

- 新增“资源监听”Tab，支持输入目标 URL、勾选常见文件后缀并启动监听任务。
- 新增 Go 与 Python 常驻 worker 通信能力，由 Go 管理任务生命周期，由 Python 使用 DrissionPage 打开浏览器并执行监听。
- 新增监听状态控制：停止监听、继续监听、结束任务；只有结束任务时才关闭浏览器。
- 新增资源列表能力：实时展示命中资源、按内容哈希去重、支持手动勾选后下载。
- 新增下载目录能力：每次任务生成 UUID 目录，下载到应用数据目录下，并支持用 Finder / VS Code 打开该目录。

## Capabilities

### New Capabilities
- `resource-monitoring`: 提供浏览器资源监听、暂停恢复、手动下载与目录打开能力。

### Modified Capabilities
- 无

## Impact

- Go 应用入口和 Wails 绑定方法会新增资源监听相关 API。
- 后端会新增任务状态管理、Python 子进程通信、下载目录管理与桌面打开目录能力。
- 前端主界面会新增一个完整 Tab 和相关状态展示、控制操作、资源表格。
- 项目会新增 Python worker 脚本和依赖说明，用于调用 DrissionPage。
