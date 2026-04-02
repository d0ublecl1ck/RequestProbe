## Context

当前项目是 `Wails + Go + React` 的单窗口桌面应用，现有功能都通过 `App` 暴露给前端。新需求同时跨越前端界面、Go 后端状态管理、Python 浏览器自动化与本地文件系统，因此属于典型的跨模块变更。

关键约束：

- 监听动作必须由 DrissionPage 驱动浏览器完成。
- Go 负责对前端暴露稳定 API，不能让前端直接理解 Python 细节。
- 停止监听时浏览器必须保留，恢复监听后只接收新的资源。
- 下载不是自动发生，而是命中后先展示列表，用户勾选后才下载。
- 下载目录固定在应用数据目录下的 `downloads/<uuid>/`，并提供 Finder / VS Code 打开入口。

参考资料中，DrissionPage 官方文档说明 `listen.start()` 必须先于页面访问；`listen.steps()` 可实时消费数据包；`listen.pause()` / `listen.resume()` / `listen.stop()` 可分别暂停、恢复和终止监听；`DataPacket.response.body` 支持返回文本、字节或字典对象。这些能力足以支持常驻监听 worker 方案。

## Goals / Non-Goals

**Goals:**

- 提供一个独立资源监听 Tab，不破坏现有请求分析流程。
- 让 Go 以单任务状态机管理 Python worker，保证开始、暂停、恢复、结束的行为一致。
- 仅对勾选后缀的资源建索引，命中结果按内容哈希去重。
- 在下载时将选中资源保存到任务 UUID 目录中，并支持桌面方式打开目录。
- 对 Python 缺失、DrissionPage 未安装、URL 非法、监听失败等场景提供明确错误反馈。

**Non-Goals:**

- 不在本次变更中实现多任务并发监听。
- 不在本次变更中实现浏览器登录态导入、代理配置或浏览器选择器。
- 不在本次变更中实现自动下载全部命中资源。
- 不在本次变更中打包完整独立 Python 运行时。

## Decisions

### 1. 采用“Go 状态机 + Python 常驻 worker”的双进程模型

选择理由：

- 用户要求停止监听时浏览器不关闭、恢复时继续使用同一个浏览器实例，这要求 Python 进程常驻并持有浏览器对象。
- Go 侧继续保持 Wails API 的统一入口，前端不需要直接处理 stdout/stderr、进程和异常细节。

备选方案：

- 每个动作都单次启动 Python 脚本：无法可靠保留浏览器和监听上下文，放弃。
- 让 Python 起 HTTP 服务：额外引入本地端口和服务编排，超出当前必要复杂度，放弃。

### 2. 使用 JSON line 双向通信协议连接 Go 和 Python

Go 启动 Python worker 后，通过 stdin 发送命令，通过 stdout 接收事件与响应，每一行都是独立 JSON 消息。

优点：

- 协议简单，易于调试和记录。
- 不依赖额外套接字、端口和 IPC 框架。
- 适合常驻 worker 和事件流。

消息模型：

```text
Go command  -> {"id":"...","type":"start_task","payload":{...}}
Py event    -> {"type":"resource_detected","payload":{...}}
Py response -> {"id":"...","type":"response","ok":true,"payload":{...}}
```

### 3. 监听结果在内存建立索引，下载在显式操作时执行

监听阶段只采集资源元信息与可下载上下文，资源列表由 Go 缓存并推送前端。用户点击“下载选中项”后，Go 再指令 Python 对指定资源执行下载。

选择理由：

- 符合“先实时列出来，再手动勾选下载”的业务要求。
- 避免所有命中资源在监听时立即落盘。
- 可在下载前做哈希去重和文件名清洗。

### 4. 去重主键采用内容哈希，URL 仅作辅助展示

同一个 URL 可能带 query 或缓存参数变化，同一内容也可能由不同 URL 提供。为了满足“要哈希去重”，以下载字节内容计算 SHA-256 作为唯一键；资源列表中同时保留 URL、后缀、状态码与 MIME 供展示。

行为定义：

- 列表去重：同一哈希只保留一条资源记录。
- 文件落盘去重：若任务目录里已存在该哈希对应文件，则跳过再次写入。

### 5. UI 采用“控制台式深色卡片 + 紧凑 launcher 按钮”

`ui-ux-pro-max` 建议本项目适合开发者工具语义：深色代码背景、运行态绿色强调、JetBrains Mono/IBM Plex Sans 风格。现有页面是浅色玻璃态，因此本次不整体翻新，而是在新 Tab 内局部引入更强的控制台视觉：

- 任务状态卡：深色底 + 绿色状态点。
- 资源表格：清晰列间距和可扫描标签。
- Finder / VS Code 打开按钮：做成与系统 launcher 接近的统一高度胶囊按钮，图标区与文字区基线对齐，保证视觉密度和点击手感。

### 6. 目录打开能力由 Go 调用本机命令完成

- Finder：`open <dir>`
- VS Code：优先 `code <dir>`，若 `code` 不存在则返回明确错误

理由：

- 用户明确要求支持 Finder / VS Code。
- 这类系统动作在 Go 层做最自然，避免 Python 和前端重复持有平台判断逻辑。

## Risks / Trade-offs

- [系统未安装 Python 或 DrissionPage] → 启动任务前做环境探测，向前端返回可读错误，避免半启动状态。
- [长时间监听导致资源列表过大] → 资源记录按哈希去重，并只保存必要元信息与下载上下文。
- [某些响应体过大或不可直接解码] → Python 下载时以字节流处理并计算哈希，不依赖文本解码。
- [VS Code CLI 未安装] → “用 VS Code 打开”失败时提示用户安装 `code` 命令或改用 Finder。
- [DrissionPage 接口行为差异] → 以官方文档中的 `listen.start()` / `listen.steps()` / `listen.pause()` / `listen.resume()` / `listen.stop()` 为最小实现闭环，避免依赖未验证能力。

## Migration Plan

1. 新增 OpenSpec 变更文档并完成实现。
2. 在项目中引入 Python worker、Go service、前端 Tab 与新绑定。
3. 本地构建验证 `wails build`，确认新绑定和前端打包通过。
4. 运行基础手工验证：开始监听、暂停、恢复、结束、打开目录、下载选中资源。

回滚策略：

- 若监听功能存在严重问题，可整体移除新增 Tab 和 Go/Python service，不影响原有请求测试功能。

## Open Questions

- 当前版本默认依赖系统 `python3` 与已安装的 DrissionPage；是否后续需要打包独立 Python 环境，可在下一次变更单独处理。
