## Context

Wails v2 提供 `options.Debug.OpenInspectorOnStartup`，该选项只在调试构建中生效。当前项目没有设置此字段，因此 `wails dev` 启动后需要开发者手动打开 DevTools。

## Decision

在 `options.App` 上增加：

```go
Debug: options.Debug{
    OpenInspectorOnStartup: true,
},
```

这样可以确保：

- `wails dev` 启动时自动打开前端 DevTools。
- 生产构建不会受影响，因为该配置只在 debug build 生效。

## Risks

- 开发时会固定弹出 DevTools，但这正是本次需求目标。
- 如果后续需要关闭该行为，可以直接把该布尔值改回 `false`。

## Verification

1. `go build ./...` 通过，说明配置字段合法。
2. 下次执行 `wails dev` 时，应用启动后会自动打开 DevTools。
