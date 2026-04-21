## Why

资源监听页面在无任务时存在不稳定的初始化行为：前端会反复触发状态更新，导致启动后控制台持续报 `Maximum update depth exceeded`，并干扰任务状态展示与交互。与此同时，Wails 调试模式下前端开发服务器地址依赖自动推断，导致 WebSocket HMR 连接可能落到 `wails.localhost` 或 `localhost:undefined` 这样的无效地址。

## What Changes

- 修正资源监听页面在“无当前任务”场景下的初始化逻辑，避免空状态下重复 `setState`
- 为资源监听任务增加前端归一化判断，确保首次进入页面时稳定显示“未开始”
- 将 Wails 调试态前端开发服务器地址改为显式回环地址，并固定 Vite HMR 主机与端口
- 为开发页补齐显式 favicon 资源，避免浏览器回退请求根路径图标时产生 404
- 清理因上述问题引发的启动控制台报错，保证开发态可稳定加载和热更新

## Capabilities

### New Capabilities

### Modified Capabilities
- `resource-monitoring`: 资源监听页面首次打开时的空任务状态展示与初始化行为将变得稳定且无副作用
- `desktop-dev-experience`: `wails dev` 启动时前端开发服务器与 HMR 连接将使用可解析且固定的地址配置

## Impact

- 影响 `frontend/src/components/resource-monitor-tab.jsx` 的初始化与状态同步逻辑
- 影响 `frontend/vite.config.mjs` 与 `wails.json` 的开发态联调配置
- 影响资源监听页首次渲染、开发态控制台输出以及热更新连接稳定性
