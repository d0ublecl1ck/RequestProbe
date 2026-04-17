## 1. OpenSpec

- [x] 1.1 为“资源监听打包自带 Python 运行时”创建独立变更提案
- [x] 1.2 记录开发态 uv 与发行态内置运行时的边界和决策
- [x] 1.3 补充资源监听规格变更，明确发行包中的运行时行为

## 2. Python 依赖治理

- [x] 2.1 为资源监听补充标准 `pyproject.toml`
- [x] 2.2 生成并提交 `uv.lock`
- [x] 2.3 将现有 `requirements-resource-monitor.txt` 迁移为受控兼容入口或移除

## 3. Go 运行时探测改造

- [x] 3.1 为发行包新增应用内置 Python 运行时定位逻辑
- [x] 3.2 调整解释器探测优先级并保留开发态回退链路
- [x] 3.3 为新探测顺序补充单元测试

## 4. 构建与打包链路

- [x] 4.1 为本地构建脚本增加 bundled runtime 生成步骤
- [x] 4.2 为 GitHub Actions release 构建增加 bundled runtime 复制步骤
- [ ] 4.3 校验 macOS、Windows、Linux 产物都包含资源监听运行时目录

## 5. 验证

- [ ] 5.1 在开发态通过 uv 环境验证资源监听
- [ ] 5.2 在不依赖系统 Python 的干净环境验证发行包资源监听可启动
- [x] 5.3 更新相关开发与发布文档
