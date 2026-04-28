## 1. OpenSpec

- [x] 1.1 补充“监听所有标签页”提案与设计说明
- [x] 1.2 为 `resource-monitoring` 增加监听范围选项的规格变更

## 2. Frontend

- [x] 2.1 在资源监听页面新增“监听所有标签页”勾选项并默认勾选
- [x] 2.2 启动任务时传递该选项，并在任务回显时同步前端状态

## 3. Backend

- [x] 3.1 扩展资源监听任务模型和 Wails 绑定，增加 `listenAllTabs`
- [x] 3.2 调整 Go 服务启动 payload，将监听范围传入 Python worker
- [x] 3.3 改造 Python worker，在多标签页模式下持续注册所有现有和新增 tab 的监听器

## 4. Verification

- [x] 4.1 运行 Go 测试验证启动 payload 与任务模型
- [x] 4.2 运行前端构建，确认新增 JSX 组件引用无缺失 import 回归
