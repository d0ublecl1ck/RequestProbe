# workspace-navigation Specification

## Purpose
TBD - created by archiving change replace-workspace-tabs-with-sidebar. Update Purpose after archive.
## Requirements
### Requirement: 桌面端必须通过左侧侧边栏切换工作区

桌面端界面 MUST 使用左侧侧边栏承载全局工作区导航，而不是顶部 Tabs。侧边栏 MUST 至少提供“字段探针”和“资源监听”两个入口，并以明确的激活态展示当前工作区。

#### Scenario: 切换到字段探针

- **WHEN** 用户点击左侧“字段探针”
- **THEN** 系统 MUST 激活字段探针工作区
- **THEN** 右侧内容区 MUST 展示请求分析相关界面

#### Scenario: 切换到资源监听

- **WHEN** 用户点击左侧“资源监听”
- **THEN** 系统 MUST 激活资源监听工作区
- **THEN** 右侧内容区 MUST 展示资源监听相关界面

### Requirement: 工作区导航改造不得引入额外全局头部

在桌面端工作区侧边栏方案下，系统 MUST 移除原有顶部品牌区和 GitHub 按钮，避免与工作区内容区重复占据头部空间。

#### Scenario: 进入主界面

- **WHEN** 用户打开主界面
- **THEN** 顶部 MUST 不再显示品牌区
- **THEN** 顶部 MUST 不再显示 GitHub 按钮
- **THEN** 工作区标题和说明 MUST 仅在当前内容区内部展示

