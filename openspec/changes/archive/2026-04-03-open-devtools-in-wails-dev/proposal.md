## Why

当前通过 `wails dev` 启动桌面应用时，前端调试器不会自动打开。开发时每次都要手动打开 DevTools，会增加前端联调和样式调试的摩擦。

## What Changes

- 在 Wails 调试构建配置中启用启动时自动打开 inspector。
- 仅影响 `wails dev` 等调试构建，不改变生产构建行为。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `desktop-dev-experience`: 改进 Wails 调试模式下的前端调试体验。

## Impact

- 修改 `main.go` 中的 Wails `options.App` 调试选项。
- 不影响前端业务逻辑、后端接口和生产打包结果。
