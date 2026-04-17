# request-lab-layout Specification

## Purpose
TBD - created by archiving change refine-request-lab-layout. Update Purpose after archive.
## Requirements
### Requirement: 字段探针桌面工作台必须优先占用可用横向空间

字段探针工作区在桌面端 MUST 使用接近视口宽度的主容器，而不是被固定在较窄的最大像素宽度内。主容器 MUST 让页面左右总留白约为 10%，以提升可用横向空间利用率。

#### Scenario: 在桌面窗口中打开字段探针

- **WHEN** 用户在桌面窗口中打开字段探针工作区
- **THEN** 主工作台容器 MUST 以接近 `90vw` 的宽度渲染
- **THEN** 左右留白 MUST 明显小于旧版固定宽度布局
- **THEN** 左侧侧边栏和右侧内容区 MUST 继续在同一行内展示

