## Context

当前前端结构存在两个核心问题：

1. `App.jsx` 同时承担应用入口、工作区路由、字段探针状态、表单渲染、结果展示和局部格式化工具函数，已经超过单组件应承担的复杂度。
2. `resource-monitor-tab.jsx` 中的打开目录交互、选项勾选行、状态卡片等已经形成可识别的局部模式，但仍以内联 JSX 形式存在。

本次重构是实现层重构，不新增业务能力，因此设计重点是组件职责重划与依赖方向校正。

## Goals / Non-Goals

**Goals:**

- 让 `App.jsx` 只保留应用壳层职责。
- 让字段探针拥有独立的工作区容器组件。
- 抽取可跨工作区复用的展示和交互组件。
- 让资源监听中的 opener 交互成为单独可测试、可替换的组件。

**Non-Goals:**

- 不修改 Go、Python 和 Wails bindings 的业务接口。
- 不引入新的状态管理库、路由层或设计系统框架。
- 不改变页面信息架构和已有交互文案。

## Decisions

### 1. 使用“壳层组件 + 工作区组件 + 共享表现组件”三级结构

新的前端结构分为三层：

```text
App Shell
├── Workspace Sidebar / Main Container
├── RequestLabWorkspace
│   ├── Form Section / Radio Card Group / Checkbox Card
│   ├── Tab Navigator
│   ├── Container Card / Code Block Card / Metric Card
│   └── Empty State
└── ResourceMonitorTab
    └── OpenerSelect
```

这样可以保证：

- 顶层只负责工作区切换。
- 字段探针把业务状态与页面编排聚合到一个领域组件。
- 共享 UI 模式向下复用，而不是让业务组件互相引用。

### 2. 保持基础 UI primitive 不重复发明

仓库已经有 `ui/button.jsx`、`ui/checkbox.jsx`、`ui/radio-group.jsx`、`ui/tabs.jsx` 等基础 primitive。本次不会再包装一套同质基础组件，而是抽取更高一层的组合组件，例如：

- `RadioCardGroup`
- `CheckboxCard`
- `WorkspaceSidebar`
- `TabNavigator`
- `ContainerCard`
- `CodeBlockCard`
- `FormSection`
- `OpenerSelect`

这样能避免“为了抽组件而抽组件”的伪抽象。

### 3. 共享组件只承载稳定展示契约，不承载业务状态机

共享组件只接收明确的 props，如 `items`、`value`、`onChange`、`actions`、`emptyState` 等，不直接耦合 Wails API、toast 或领域状态对象。业务状态机仍留在 `RequestLabWorkspace` 与 `ResourceMonitorTab` 中。

### 4. 优先抽取高重复、高信息密度的 UI 片段

本次优先拆分以下片段：

- 工作区侧边栏 item 组
- 字段探针右侧 Tab Navigator
- 字段探针卡片容器与代码块区域
- 字段探针表单 Section
- 资源监听 opener 选择器

不强行把所有表格或统计块都独立成单文件，避免过度碎片化。

### 5. 修正字段探针输入卡片的滚动高度链路

字段探针输入面板在重构后采用 `ContainerCard -> CardContent -> ScrollArea -> CardFooter` 结构。滚动是否生效依赖 `CardContent` 成为可收缩的纵向 flex 容器，否则 `ScrollArea` 的 `flex-1` 无法获得剩余高度。

因此该面板必须满足：

- `CardContent` 使用 `flex min-h-0 flex-1 flex-col`
- `ScrollArea` 使用 `min-h-0 flex-1`

这样才能保证底部操作区固定在卡片内，长内容在上方滚动区内部滚动，而不是把整张卡片继续撑高。

### 6. 禁止页面最外层参与垂直滚动

即使内部卡片不再真正滚动，如果 `html`、`body`、`#app` 或应用壳层仍允许页面级 overscroll，macOS 触控板仍会出现上下回弹动画，造成“看起来还能滚”的错误反馈。

因此最外层必须满足：

- `html`、`body`、`#app` 统一为 `height: 100%`
- `html`、`body`、`#app` 禁止页面级 `overflow`
- 最外层应用壳层禁用垂直 `overscroll-behavior`

这样可以把滚动责任严格限制在显式的内部滚动容器上。

### 7. 放宽主工作台的最大宽度占比

原先应用壳层使用固定 `max-w-[1440px]`，在 1920px 左右的桌面屏幕上会留下过大的左右留白，导致主工作区显得过窄。

因此主工作台容器改为：

- 以大约 `90vw` 为目标宽度
- 在超宽屏上设置合理上限，避免无限拉伸

这样可以让左右总留白接近 10%，同时继续保持居中和桌面端可读性。

### 8. 对所有页面与关键组件实施宽高响应式约束

当前项目虽然是桌面应用，但窗口宽高仍会频繁变化，因此不能依赖固定 `100vh` / `calc(...)` 高度和恒定双栏布局。

本次统一采用以下原则：

- 应用壳层使用 `min-h-0` 的 flex 高度分配，而不是多个独立固定高度
- 字段探针与资源监听在宽度不足时允许纵向堆叠
- 在桌面宽度下恢复双栏，并把滚动责任下沉到各自面板
- 长内容区使用内部 `overflow-y-auto` 或 `ScrollArea`
- 侧边栏在较窄宽度下改为顶部横向/网格导航，而不是持续占用固定左列宽度

这样可以同时覆盖：

- 较矮窗口下的底部溢出
- 较窄窗口下的左右挤压
- 表格和长代码块导致的内容区失衡

## Risks / Trade-offs

- 文件数量会增加，但单文件复杂度显著下降，这是值得的结构性交换。
- 如果共享组件抽象过度，会导致 props 过宽、可读性下降；因此只抽当前已经稳定的 UI 模式。
- 由于当前仓库没有完整前端测试基建，本次主要依赖构建校验与语法打包校验来验证重构安全性。

## Verification

1. `App.jsx` 不再包含字段探针的大段页面 JSX。
2. 字段探针页面能够通过独立工作区组件渲染。
3. 资源监听中的 Finder / VS Code 打开器通过独立组件渲染。
4. 前端构建或等价打包校验通过，证明重构未破坏模块依赖。
5. 字段探针存在长内容时，请求输入面板能够在卡片内部产生垂直滚动。
6. 页面最外层不再出现垂直滚动和回弹动画。
7. 在桌面宽屏下，主工作台容器宽度约占视口 90%，不再出现过大的左右留白。
8. 在较窄宽度或较矮高度下，字段探针页、资源监听页和侧边栏都不会出现底部溢出或固定双栏挤压。
