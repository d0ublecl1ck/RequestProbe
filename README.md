<img width="2940" height="1726" alt="image" src="https://github.com/user-attachments/assets/7789698c-f68e-46ee-afa6-80fc34bcf80f" />

## 资源监听 Python 环境

开发态使用 `uv` 管理资源监听所需的 Python 依赖：

```bash
UV_PROJECT_ENVIRONMENT=.venv-monitor uv sync --python 3.13 --no-dev
```

完成后，资源监听会优先使用仓库内的 `.venv-monitor`。

发行构建会在打包阶段把资源监听所需的 Python runtime 一起整理进最终产物，避免最终用户手动安装 `Python + DrissionPage`。
