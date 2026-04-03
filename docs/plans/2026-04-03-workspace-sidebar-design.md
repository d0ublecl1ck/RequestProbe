# Workspace Sidebar Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将主界面的两个工作区入口从顶部 Tabs 改为左侧侧边栏导航，并移除顶部品牌区和 GitHub 按钮。

**Architecture:** 保留 `activeWorkspaceTab` 作为唯一状态源，只替换导航呈现和外层布局容器。两个工作区内部内容尽量不动，减少对现有交互和业务逻辑的影响。

**Tech Stack:** React, Vite, Wails, Tailwind 风格工具类

---

### Task 1: 调整主布局骨架

**Files:**
- Modify: `frontend/src/App.jsx`

**Step 1:** 删除顶部品牌区与 GitHub 按钮相关 JSX 和图标引用。  
**Step 2:** 以左侧侧边栏和右侧内容区重组 `Tabs` 外层布局。  
**Step 3:** 保留 `activeWorkspaceTab` 状态，不改动工作区切换逻辑。  

### Task 2: 替换导航呈现

**Files:**
- Modify: `frontend/src/App.jsx`

**Step 1:** 移除顶部 `TabsList` 和 `TabsTrigger`。  
**Step 2:** 新增左侧按钮式工作区导航并绑定 `setActiveWorkspaceTab`。  
**Step 3:** 为当前导航项提供明确高亮样式。  

### Task 3: 验证

**Files:**
- Modify: `openspec/changes/replace-workspace-tabs-with-sidebar/tasks.md`

**Step 1:** 运行 `npm run build`（在 `frontend/` 下）。  
**Step 2:** 如构建通过，勾选验证任务。  
**Step 3:** 运行 `openspec validate replace-workspace-tabs-with-sidebar`。  
