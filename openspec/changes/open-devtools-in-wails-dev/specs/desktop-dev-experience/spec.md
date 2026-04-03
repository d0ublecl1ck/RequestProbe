## ADDED Requirements

### Requirement: Wails 调试模式必须支持自动打开前端 DevTools

当开发者通过 Wails 调试模式启动桌面应用时，系统 MUST 自动打开前端 DevTools，以减少手动调试成本。

#### Scenario: 使用 wails dev 启动应用

- **WHEN** 开发者执行 `wails dev`
- **THEN** 系统 MUST 在应用启动后自动打开前端 DevTools
- **THEN** 该行为 MUST 仅作用于调试构建，不影响生产构建
