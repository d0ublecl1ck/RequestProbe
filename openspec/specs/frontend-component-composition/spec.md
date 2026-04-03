# frontend-component-composition Specification

## Purpose
TBD - created by archiving change extract-shared-frontend-components. Update Purpose after archive.
## Requirements
### Requirement: 主界面必须通过分层 React 组件组合工作区界面

前端主界面 MUST 将应用壳层、工作区页面和共享展示组件分层组织，而不是把字段探针与资源监听的主要界面长期堆叠在单个入口组件中。

#### Scenario: 渲染字段探针工作区

- **WHEN** 用户进入“字段探针”工作区
- **THEN** 系统 MUST 通过独立的字段探针工作区组件渲染页面主体
- **THEN** 输入表单、结果 Tab 导航、代码块与卡片容器 MUST 由可复用子组件组合完成

#### Scenario: 字段探针输入面板出现长内容

- **WHEN** 字段探针输入面板中的表单内容高度超过卡片可视区域
- **THEN** 系统 MUST 在卡片内部提供垂直滚动
- **THEN** 底部操作区 MUST 保持在卡片底部而不是随内容整体被撑出

#### Scenario: 用户在应用最外层继续向上或向下滚动

- **WHEN** 页面根层已经没有可滚动内容
- **THEN** 系统 MUST 不允许最外层容器产生垂直滚动
- **THEN** 系统 MUST 不再出现页面级上下回弹动画

#### Scenario: 在桌面宽屏下查看主工作台

- **WHEN** 用户在桌面宽屏环境中打开应用
- **THEN** 主工作台容器 MUST 占据大约 90% 的视口宽度
- **THEN** 页面左右外边距 MUST 收敛到约 10% 总量而不是保留过大的空白

#### Scenario: 在较窄宽度下查看工作区

- **WHEN** 应用窗口宽度不足以稳定承载固定双栏布局
- **THEN** 系统 MUST 允许工作区切换为纵向堆叠布局
- **THEN** 侧边栏导航 MUST 不再强占固定左侧宽度

#### Scenario: 在较矮高度下查看长内容页面

- **WHEN** 应用窗口高度降低且页面存在长表单、长代码块或长表格
- **THEN** 系统 MUST 通过内部滚动区域承载长内容
- **THEN** 页面 MUST 不出现底部内容被挤出容器的问题

#### Scenario: 渲染资源监听工作区打开器

- **WHEN** 用户进入“资源监听”工作区并查看下载目录打开入口
- **THEN** 系统 MUST 通过独立 opener 组件渲染 Finder / VS Code 打开方式选择器
- **THEN** opener 组件 MUST 与资源监听业务动作解耦，只通过显式 props 接收选项、当前值和打开动作

