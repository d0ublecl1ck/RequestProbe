## 1. OpenSpec

- [x] 1.1 新增“顶部工作区切换改为左侧侧边栏”的变更提案
- [x] 1.2 记录布局方案、边界和验证标准

## 2. 前端改造

- [x] 2.1 删除 `App.jsx` 顶部品牌区和 GitHub 按钮
- [x] 2.2 将顶部 Tabs 导航改为左侧侧边栏导航
- [x] 2.3 保持两个工作区内容逻辑不变并接入新的布局容器

## 3. 验证与归档

- [ ] 3.1 运行前端构建验证改造后的布局入口可编译
- [ ] 3.2 更新任务状态并准备归档当前 OpenSpec 变更

## Notes

- `npm run build` 当前被仓库缺失的 `wailsjs` 生成目录阻塞，不是本次布局改造引起的 JSX 错误。
- 已使用 `npx esbuild src/App.jsx --bundle --format=esm --loader:.jsx=jsx '--external:../wailsjs/*' '--external:../../wailsjs/*'` 验证本次前端改动的语法与模块结构可被打包。
