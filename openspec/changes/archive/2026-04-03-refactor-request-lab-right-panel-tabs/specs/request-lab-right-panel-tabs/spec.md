## ADDED Requirements

### Requirement: 用户可以通过右侧 Tab 分栏查看字段探针结果
系统 MUST 将字段探针页面右侧结果区组织为固定 Tab 分栏，而不是将四类结果长期纵向堆叠在同一滚动列中。

#### Scenario: 默认打开 Python 代码
- **WHEN** 用户进入字段探针页面且右侧结果区首次渲染
- **THEN** 系统 MUST 默认选中 `Python 代码` Tab
- **THEN** 系统 MUST 在右侧区提供 `Python 代码`、`请求测试结果`、`测试摘要`、`简化代码` 四个固定 Tab

#### Scenario: 用户切换查看不同结果
- **WHEN** 用户点击任一右侧结果 Tab
- **THEN** 系统 MUST 切换到对应结果面板
- **THEN** 系统 MUST 保留该面板原有内容、操作按钮和空状态提示

### Requirement: 字段分析完成后界面自动聚焦测试摘要
系统 MUST 在字段分析成功完成并生成摘要结果后，将右侧结果区自动切换到 `测试摘要` Tab。

#### Scenario: 字段分析成功后自动跳转摘要
- **WHEN** 用户触发字段分析且系统成功返回字段分析结果
- **THEN** 系统 MUST 自动选中 `测试摘要` Tab
- **THEN** 系统 MUST 展示最新一次字段分析对应的摘要内容

#### Scenario: 非字段分析动作不强制跳转摘要
- **WHEN** 用户仅执行请求解析、单次请求测试或手动切换其他右侧结果 Tab
- **THEN** 系统 MUST NOT 因这些动作自动切换到 `测试摘要` Tab
