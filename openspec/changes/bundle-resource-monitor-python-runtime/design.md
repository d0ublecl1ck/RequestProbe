## Context

当前项目是 `Wails + Go + React + Python worker` 的桌面应用。前端静态资源通过 Go `embed` 打包进程序，资源监听的 Python worker 脚本也通过 `embed` 内嵌，但 Python 解释器本身和 `DrissionPage` 依赖没有被纳入构建产物。

现有实现的解释器查找顺序是：

1. `REQUESTPROBE_PYTHON`
2. 仓库根目录下 `.venv-monitor` / `.venv`
3. 宿主机 `python3` / `python`

这对于开发机尚可，但对发行包无效，因为发行包不具备仓库目录，也不应要求用户自行提供系统级 Python 环境。

## Goals / Non-Goals

**Goals:**

- 让资源监听在发行版中默认可用，不依赖宿主机预装 Python。
- 保留开发态用 `uv` 管理依赖与虚拟环境的工作流。
- 让解释器查找顺序清晰、可诊断、可覆盖。
- 保持 Python worker 通信协议不变，降低改造范围。

**Non-Goals:**

- 不在本次变更中重写 Python worker 为 Go 原生实现。
- 不引入首次启动在线下载 Python 依赖的流程。
- 不要求把 uv 本身作为最终用户运行时的一部分长期驻留。

## Options Considered

### 方案 A: 直接把 `.venv-monitor` 打进安装包

**优点:**

- 实现看似简单，开发环境和发行环境表面一致。

**缺点:**

- `venv` 对创建路径和宿主环境有隐式耦合，可移植性差。
- CI 机器生成的虚拟环境直接复制到终端用户机器，稳定性不可控。
- 不同平台、不同架构无法复用。

**结论:**

- 放弃。该方案短期省事，长期维护成本最高。

### 方案 B: 发行包不带运行时，首次启动时自动创建 `.venv-monitor`

**优点:**

- 安装包较小。
- 改动范围相对可控。

**缺点:**

- 仍依赖宿主机已有 Python。
- 仍可能依赖网络拉取依赖。
- 首次启动慢且失败面大，不符合桌面应用分发预期。

**结论:**

- 放弃。它只是把安装成本从打包阶段转嫁给最终用户。

### 方案 C: 用 uv 锁依赖，构建时整理“应用私有 Python 运行时”

**优点:**

- 开发态和构建态都由 uv 统一依赖来源。
- 发行包可脱离系统 Python 运行资源监听。
- 解释器来源明确，可优先选择内置运行时。

**缺点:**

- 构建流程更复杂，安装包会变大。
- 需要分别处理 macOS、Windows、Linux 的运行时目录布局。

**结论:**

- 采用。这是当前上下文下最稳、最可维护的方案。

## Chosen Design

### 1. 依赖源统一为 uv 项目

新增标准 Python 项目元数据，使 `DrissionPage` 及其运行依赖由 `uv` 锁定和安装。开发态继续支持：

- `uv sync`
- `uv run ...`
- `repo/.venv-monitor`

但发行态不再直接依赖该虚拟环境目录。

### 2. 构建时生成“bundled runtime”目录

在每个平台的打包过程中，生成一个应用私有运行时目录，并随最终产物一同分发。

```text
macOS:
RequestProbe.app
└── Contents/Resources/python/
    ├── bin/python3
    └── lib/pythonX.Y/site-packages/...

Windows/Linux:
build/bin/python/
├── python(.exe)
└── Lib|lib/.../site-packages/...
```

### 3. Go 侧解释器查找顺序调整

解释器查找顺序调整为：

1. `REQUESTPROBE_PYTHON`
2. 应用包内置运行时
3. 开发态 `.venv-monitor` / `.venv`
4. 系统 `python3` / `python`

这样能同时覆盖：

- 调试场景：显式指定解释器
- 开发场景：仓库内 uv 环境
- 分发场景：应用内置运行时

### 4. 继续复用内嵌 Python worker 脚本

`resource_monitor_worker.py` 仍然通过 Go `embed` 写入缓存目录再执行，不把 worker 文件单独散落到发行包外层目录。这样只需替换解释器来源，不需要重做 IPC 协议或文件发现逻辑。

## Build Flow

```text
uv lock / uv sync
        |
        v
整理平台私有 Python runtime
        |
        v
wails build
        |
        v
将 runtime 复制到最终 app / bin
        |
        v
压缩或制作安装包
```

## Risks / Trade-offs

- 安装包体积增加：
  这是换取“开箱即用”的必要成本。
- 各平台运行时目录差异：
  需要把运行时定位逻辑做成显式、可测试的路径探测，而不是硬编码单一路径。
- Python 版本与 DrissionPage 兼容性：
  需要在 `uv` 锁文件和 CI 构建中固定可用组合，避免构建机漂移。

## Verification

1. 在干净机器或干净虚拟机上安装产物，不额外安装 Python。
2. 启动应用并创建资源监听任务，应成功拉起浏览器而不是报环境缺失。
3. 开发态在仓库根目录运行时，`.venv-monitor` 仍可被正常探测并工作。
4. 显式设置 `REQUESTPROBE_PYTHON` 时，应优先采用指定解释器。
