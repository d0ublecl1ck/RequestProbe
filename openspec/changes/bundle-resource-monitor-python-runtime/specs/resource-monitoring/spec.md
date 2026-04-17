## ADDED Requirements

### Requirement: 资源监听在桌面发行包中默认自带可用运行时
系统 MUST 在桌面发行包中提供资源监听所需的 Python 运行时与 `DrissionPage` 依赖，使最终用户在未预装系统 Python 的情况下也可以启动资源监听。

#### Scenario: 发行包在无系统 Python 的机器上启动资源监听
- **WHEN** 用户在仅安装了 RequestProbe 的机器上打开资源监听页面并点击“开始监听”
- **THEN** 系统 MUST 优先使用应用内置的 Python 运行时启动资源监听 worker
- **THEN** 系统 MUST NOT 要求用户额外安装 Python 或 DrissionPage

### Requirement: 开发态与手动指定解释器仍然可用
系统 MUST 同时保留开发态 uv 环境与显式指定解释器的支持，以便本地调试、CI 验证与故障排查。

#### Scenario: 显式设置 REQUESTPROBE_PYTHON
- **WHEN** 运行环境设置了 `REQUESTPROBE_PYTHON`
- **THEN** 系统 MUST 优先使用该解释器启动资源监听 worker

#### Scenario: 开发态使用仓库内 uv 环境
- **WHEN** 用户在仓库根目录运行桌面应用且存在 `.venv-monitor` 或 `.venv`
- **THEN** 系统 MUST 将仓库内可用解释器作为应用内置运行时之后的回退选项
- **THEN** 资源监听 MUST 可以在该开发环境中正常启动
