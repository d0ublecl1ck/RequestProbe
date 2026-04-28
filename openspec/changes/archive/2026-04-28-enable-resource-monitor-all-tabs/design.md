## Context

DrissionPage 当前版本的监听器是页面/标签页对象级能力，不会自动跨浏览器全部标签页共享。现有 worker 只有一个 `self.page.listen`，因此只能消费一个标签页的数据。要支持“监听所有标签页”，必须在同一浏览器会话里主动枚举 tab，并给每个 tab 各自启动 `listen.start()`。

## Goals / Non-Goals

**Goals:**
- 提供一个可显式控制的“监听所有标签页”启动选项
- 默认覆盖浏览器中所有现有和后续新增标签页
- 保持暂停、恢复、结束任务时对所有活动监听器统一生效

**Non-Goals:**
- 不改变资源去重和请求记录结构
- 不为单个 tab 提供更细粒度的筛选规则
- 不引入新的独立浏览器进程或第二套 worker

## Decisions

### 决策 1：用布尔勾选项暴露监听范围，并默认开启

前端新增 `listenAllTabs` 布尔状态，默认值为 `true`。任务启动时将该值传递给 Go 和 Python，当前任务也会持久携带该值，便于页面刷新后正确回显。

选择原因：
- 行为清晰，不会悄悄改变老用户的预期
- 默认值符合“我想监控整个会话”的更常见意图

### 决策 2：Python worker 在多标签页模式下维护 tab 监听注册表

worker 增加 `listen_all_tabs` 和 `tab_listeners`。启动任务后先注册当前浏览器已有 tab，再在监听循环中轮询 `tab_ids`，发现新 tab 就补注册，发现已关闭 tab 就清理对应引用。

选择原因：
- 符合 DrissionPage 监听器的 tab 级能力边界
- 不需要额外浏览器实例，仍复用现有任务生命周期

备选方案：
- 继续只监听 `ChromiumPage`，期待其天然覆盖所有标签页。未采用，因为官方文档说明监听器挂在具体页面对象上，当前实现也已验证只覆盖单 tab。

### 决策 3：暂停、恢复、结束任务时统一操作所有已注册监听器

在多标签页模式下，`pause_task` / `resume_task` / `end_task` 不再只对单个 `self.page.listen` 操作，而是遍历所有已注册 tab listener 统一执行；单标签页模式保持原有路径。

选择原因：
- 让任务生命周期与监听范围保持一致
- 避免部分 tab 仍在产生日志，状态却显示“已暂停”

## Risks / Trade-offs

- [风险] 新标签页创建到注册之间存在极短时间窗，可能漏掉刚打开瞬间的最早几个请求
  - 缓解：在监听循环中高频同步 `tab_ids`，并在发现新 tab 后立即启动监听
- [风险] 已关闭 tab 的 listener 调用可能抛异常
  - 缓解：清理失效 tab 引用，并在单个 tab 级别吞掉关闭态异常，不影响其它 tab

## Migration Plan

- 增加 OpenSpec proposal / design / tasks / spec delta
- 扩展前端、Go、Python 参数链路
- 增加 Go 单元测试验证默认值与启动 payload
- 运行 Go 测试与前端构建

## Open Questions

- 当前不把 `tab_id` 单独暴露到前端请求列表；若后续需要按标签页筛选，再作为独立增强处理
