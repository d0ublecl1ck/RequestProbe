# desktop-dev-experience Specification

## Purpose
TBD - created by archiving change open-devtools-in-wails-dev. Update Purpose after archive.
## Requirements
### Requirement: Wails 调试模式必须支持自动打开前端 DevTools

当开发者通过 Wails 调试模式启动桌面应用时，系统 MUST 自动打开前端 DevTools，以减少手动调试成本。

#### Scenario: 使用 wails dev 启动应用

- **WHEN** 开发者执行 `wails dev`
- **THEN** 系统 MUST 在应用启动后自动打开前端 DevTools
- **THEN** 该行为 MUST 仅作用于调试构建，不影响生产构建

### Requirement: Wails 调试模式必须使用稳定且可解析的前端开发服务器地址

当开发者通过 `wails dev` 启动桌面应用时，系统 MUST 使用显式且可解析的本地回环地址连接前端开发服务器与 HMR WebSocket，而不是依赖 WebView 运行时的隐式主机名推断。

#### Scenario: 使用 wails dev 启动资源监听页面
- **WHEN** 开发者执行 `wails dev` 并打开桌面应用
- **THEN** 前端资源 MUST 从显式配置的本地开发服务器地址加载
- **THEN** HMR WebSocket MUST 使用有效主机和端口建立连接
- **THEN** 控制台 MUST NOT 出现 `wails.localhost` 无法解析或 `localhost:undefined` 之类的 HMR 地址错误
- **THEN** 控制台 MUST NOT 因缺失默认 favicon 资源而出现 `GET /favicon.ico 404`

