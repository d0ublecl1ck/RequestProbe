import React, { useState } from 'react';
import { Toaster } from 'sonner';

import { WorkspaceSidebar } from './components/layout/workspace-sidebar.jsx';
import { RequestLabWorkspace } from './components/request-lab/request-lab-workspace.jsx';
import { ResourceMonitorTab } from './components/resource-monitor-tab.jsx';
import { SettingsTab } from './components/settings-tab.jsx';
import { TooltipProvider } from './components/ui/tooltip.jsx';
import { Tabs, TabsContent } from './components/ui/tabs.jsx';

const workspaceItems = [
  {
    value: 'request-lab',
    label: '字段探针',
    subtitle: '解析请求、测试字段必要性、生成 Python 代码',
  },
  {
    value: 'resource-monitor',
    label: '资源监听',
    subtitle: '启动浏览器监听任务，采集资源与请求',
  },
  {
    value: 'settings',
    label: '设置',
    subtitle: '维护资源监听默认保存位置与全局选项',
  },
];

export default function App() {
  const [activeWorkspaceTab, setActiveWorkspaceTab] = useState('request-lab');
  const activeWorkspaceItem = workspaceItems.find((item) => item.value === activeWorkspaceTab) || workspaceItems[0];

  return (
    <TooltipProvider>
      <div className="app-shell">
        <Toaster richColors position="top-right" />
        <div className="app-workspace-frame">
          <Tabs
            value={activeWorkspaceTab}
            onValueChange={setActiveWorkspaceTab}
            className="workspace-shell"
          >
            <WorkspaceSidebar
              items={workspaceItems}
              value={activeWorkspaceTab}
              onChange={setActiveWorkspaceTab}
              currentTitle={activeWorkspaceItem.label}
              currentDescription={activeWorkspaceItem.subtitle}
            />

            <div className="workspace-main min-h-0 min-w-0 flex-1 overflow-hidden">
              <TabsContent value="request-lab" className="mt-0 h-full overflow-hidden">
                <RequestLabWorkspace />
              </TabsContent>

              <TabsContent value="resource-monitor" className="mt-0 h-full overflow-hidden">
                <ResourceMonitorTab />
              </TabsContent>

              <TabsContent value="settings" className="mt-0 h-full overflow-hidden">
                <SettingsTab />
              </TabsContent>
            </div>
          </Tabs>
        </div>
      </div>
    </TooltipProvider>
  );
}
