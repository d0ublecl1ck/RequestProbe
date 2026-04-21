# JSX 新增组件后遗漏 import 事故记录

## 背景

在前端重构过程中，`RequestLabWorkspace` 新增了 `Badge` 组件用于页面摘要状态展示，但文件顶部没有同步补充：

```jsx
import { Badge } from '../ui/badge.jsx';
```

结果是 Wails 开发态在运行时直接报错：

```text
ReferenceError: Can't find variable: Badge
```

页面因此在 React 渲染阶段崩溃，字段探针工作区无法正常打开。

## 错误本质

这不是业务逻辑错误，而是典型的 JSX 符号引用错误：

- 在 JSX 中使用了新的组件标识符
- 但没有在当前模块显式导入该标识符
- 构建或运行时才暴露为 `ReferenceError`

## 为什么会发生

- 重构时关注了页面结构和视觉层级，遗漏了最基本的模块依赖检查
- 修改集中在 JSX 返回结构，新增符号时没有做“引用即导入”的收口检查
- 没有在改动后第一时间做一次针对新增 JSX 标识符的人工核对

## 直接修复

在 [request-lab-workspace.jsx](/Users/d0ublecl1ck/RequestProbe/frontend/src/components/request-lab/request-lab-workspace.jsx) 顶部补充：

```jsx
import { Badge } from '../ui/badge.jsx';
```

随后重新执行前端构建，确认错误消失。

## 强制预防规则

以后凡是修改 React / JSX 文件，尤其是重构返回结构、补页面摘要区、插入新组件时，必须执行下面检查：

1. 每新增一个 JSX 组件标签，立即确认当前文件顶部已有对应 import。
2. 完成 JSX 改动后，人工扫描一遍当前文件中所有新增的大写组件名，逐个核对 import。
3. 在交付前至少跑一次前端构建，不能只依赖肉眼检查。
4. 把“缺少 import 导致运行时崩溃”视为低级错误，禁止重复出现。

## 适用范围

- `frontend/src/**/*.jsx`
- `frontend/src/**/*.tsx`
- 任意包含 JSX 返回结构的 React 模块
