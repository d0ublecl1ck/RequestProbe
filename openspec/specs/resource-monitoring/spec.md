# resource-monitoring Specification

## Purpose
TBD - created by archiving change add-resource-monitor-tab. Update Purpose after archive.
## Requirements
### Requirement: 用户可以创建资源监听任务
系统 MUST 提供一个资源监听入口，允许用户输入目标 URL、选择需要监听的常见文件后缀，并启动一个新的浏览器监听任务。启动任务时系统 MUST 生成唯一任务 UUID，并为该任务分配应用数据目录下的独立下载文件夹。

#### Scenario: 成功启动监听任务
- **WHEN** 用户在资源监听页输入合法 URL、勾选至少一个文件后缀并点击“开始监听”
- **THEN** 系统创建一个新的任务 UUID
- **THEN** 系统在应用数据目录下创建 `downloads/<uuid>/` 目录
- **THEN** 系统通过 Python worker 打开浏览器并访问该 URL
- **THEN** 系统将任务状态更新为“监听中”

#### Scenario: 启动时输入非法
- **WHEN** 用户未输入 URL 或输入非法 URL 就点击“开始监听”
- **THEN** 系统 MUST 阻止启动
- **THEN** 系统 MUST 向界面返回明确错误信息

### Requirement: 用户可以暂停、恢复和结束监听任务
系统 MUST 支持对当前监听任务执行暂停、恢复和结束操作。暂停时 MUST 保留浏览器实例且不再接收新的监听结果；恢复时 MUST 继续使用同一个浏览器实例并且只接收恢复后的新资源；结束时 MUST 终止监听并关闭浏览器。

#### Scenario: 暂停监听
- **WHEN** 当前任务处于“监听中”且用户点击“停止监听”
- **THEN** 系统 MUST 调用监听暂停能力停止接收新资源
- **THEN** 系统 MUST 保留当前浏览器实例
- **THEN** 系统 MUST 将任务状态更新为“已暂停”

#### Scenario: 恢复监听
- **WHEN** 当前任务处于“已暂停”且用户点击“继续监听”
- **THEN** 系统 MUST 在原浏览器实例上恢复监听
- **THEN** 系统 MUST 只接收恢复之后新出现的资源
- **THEN** 系统 MUST 将任务状态更新为“监听中”

#### Scenario: 结束任务
- **WHEN** 用户点击“结束任务”
- **THEN** 系统 MUST 停止监听并关闭浏览器
- **THEN** 系统 MUST 将任务状态更新为“已结束”
- **THEN** 已结束任务 MUST 不可再次恢复

### Requirement: 系统可以实时展示并按哈希去重资源
系统 MUST 将符合所选文件后缀的资源实时展示到资源列表中。资源列表 MUST 按内容哈希去重，而不是仅按 URL 去重。每条资源记录 MUST 至少包含 URL、后缀、状态信息、哈希标识、下载状态与勾选状态。

#### Scenario: 命中资源被展示
- **WHEN** 监听中的浏览器接收到一个符合已选文件后缀的资源响应
- **THEN** 系统 MUST 将该资源加入资源列表
- **THEN** 界面 MUST 在无需刷新页面的情况下展示该资源

#### Scenario: 重复资源被哈希去重
- **WHEN** 监听过程中再次出现内容哈希相同的资源
- **THEN** 系统 MUST 不新增重复列表项
- **THEN** 系统 MAY 更新已有记录的最近命中时间或来源 URL

### Requirement: 用户可以手动勾选资源并下载到任务目录
系统 MUST 支持用户在资源列表中手动勾选若干资源并执行下载。下载时系统 MUST 将文件保存到当前任务 UUID 对应目录，并继续使用内容哈希避免重复落盘。

#### Scenario: 下载选中资源
- **WHEN** 用户勾选一个或多个尚未下载的资源并点击“下载选中项”
- **THEN** 系统 MUST 将选中资源保存到当前任务的 UUID 目录
- **THEN** 系统 MUST 为每个成功保存的资源更新下载状态
- **THEN** 系统 MUST 对已存在相同内容哈希的文件跳过重复写入

#### Scenario: 未勾选资源时下载
- **WHEN** 用户未勾选任何资源就点击“下载选中项”
- **THEN** 系统 MUST 阻止下载动作
- **THEN** 系统 MUST 向界面返回明确提示

### Requirement: 用户可以用桌面应用打开任务目录
系统 MUST 在资源监听页面提供打开当前任务下载目录的快捷按钮，至少支持 Finder 和 VS Code 两种方式。

#### Scenario: 用 Finder 打开目录
- **WHEN** 当前任务目录已存在且用户点击“Finder”
- **THEN** 系统 MUST 使用系统默认 Finder 打开该目录

#### Scenario: 用 VS Code 打开目录
- **WHEN** 当前任务目录已存在且用户点击“VS Code”
- **THEN** 系统 MUST 尝试调用 VS Code 命令行打开该目录
- **THEN** 若本机未安装 `code` 命令，系统 MUST 返回明确错误而不是静默失败

