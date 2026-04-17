## Why

当前资源监听能力依赖目标机器预先具备可用的 `Python + DrissionPage` 运行环境。开发机可以通过本地 `.venv-monitor` 临时解决，但现有 `wails build` 产物不会自动携带 Python 解释器、`DrissionPage` 依赖或 uv 创建的环境，因此安装包发到其他机器后，资源监听能力仍然可能在首次使用时直接失败。

这使得“桌面应用已安装即可使用资源监听”这一用户预期无法成立，也让构建结果与运行结果高度依赖宿主机状态，缺乏可重复性和可分发性。

## What Changes

- 将资源监听的 Python 依赖管理统一为 `uv` 项目工作流，而不是仅靠独立 `requirements` 文件和手工虚拟环境。
- 为桌面发行包增加“应用私有 Python 运行时”打包步骤，使资源监听在发行环境中默认不依赖系统 Python。
- 调整 Go 侧 Python 解释器探测顺序，优先使用发行包内置运行时，其次才回退到显式环境变量和开发态环境。
- 更新构建脚本与 CI，使 macOS、Windows、Linux 构建产物都能包含资源监听所需的最小 Python 运行时目录。

## Capabilities

### New Capabilities

- `resource-monitoring-runtime`: 提供资源监听在桌面发行包中自带 Python 运行时并可直接启动的能力。

### Modified Capabilities

- `resource-monitoring`: 资源监听启动时的运行时来源从“仅依赖外部环境”扩展为“优先使用内置运行时，开发态和手动指定环境作为回退”。

## Impact

- Python 依赖声明会从零散文件升级为标准 `uv` 项目元数据。
- Go 后端的 Python 解释器查找逻辑会新增对应用资源目录的识别。
- GitHub Actions 与本地打包脚本会新增运行时整理与复制步骤。
- 安装包体积会增加，但资源监听能力将不再要求最终用户手动安装 Python 或 DrissionPage。

## Decision Snapshot

```text
开发态:
repo/.venv-monitor  <- uv sync / uv run

发行态:
RequestProbe.app or build/bin
└── bundled python runtime
    ├── python executable
    └── DrissionPage + deps
```
