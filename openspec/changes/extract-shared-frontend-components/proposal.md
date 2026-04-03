## Why

当前前端主界面把工作区壳层、字段探针业务状态、表单编排、代码展示、分析结果展示和资源监听局部交互混在 `frontend/src/App.jsx` 与 `frontend/src/components/resource-monitor-tab.jsx` 中。单文件过大导致几个问题：

- 组件职责混杂，布局层与业务层难以独立维护。
- 多处卡片、选项组、代码块、空态和导航结构重复表达，后续修改成本高。
- `resource-monitor-tab.jsx` 中的打开器选择器是独立交互单元，却没有被提炼为可复用组件。

本次需要在不改变现有业务能力的前提下，抽取稳定的 React 组件边界，降低耦合并提升全局复用性。

## What Changes

- 将字段探针页面从 `App.jsx` 中拆出为独立工作区组件，只保留应用壳层与工作区切换逻辑在顶层。
- 抽取主容器、侧边栏分组、Tab Navigator、Container Card、Code Block Card、Form Section、Radio 卡片组选项、Checkbox 卡片选项等前端复用组件。
- 将资源监听中的 Finder / VS Code 打开器下拉交互抽取为独立 `OpenerSelect` 组件。
- 修正字段探针“请求输入”卡片的滚动布局链路，保证长表单内容在卡片内部可滚动。
- 禁止页面最外层容器产生垂直滚动与 overscroll 回弹，只保留内部滚动区处理长内容。
- 放宽应用主工作台容器宽度，占用约 90% 视口宽度，仅保留约 10% 的总外边距。
- 全面收敛应用壳层、字段探针页、资源监听页和共享组件的宽高响应式行为，确保窄宽度和矮高度下不会出现布局溢出。
- 保持现有 Wails 调用、字段探针状态流转、资源监听状态流转与既有功能行为不变。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- 无

## Impact

- `frontend/src/App.jsx` 将收敛为应用入口壳层和工作区切换容器。
- 字段探针与资源监听的 UI 结构会改为组合式组件树，但不改动后端接口或业务语义。
- 本次变更主要影响 React 前端文件结构、组件依赖关系和复用边界，不影响 Go、Python 与 Wails 绑定接口。
