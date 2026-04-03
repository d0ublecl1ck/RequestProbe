import React, { useState } from 'react';
import { Github } from 'lucide-react';
import { Toaster } from 'sonner';

import { MainContainer } from './components/layout/main-container.jsx';
import { WorkspaceSidebar } from './components/layout/workspace-sidebar.jsx';
import { RequestLabWorkspace } from './components/request-lab/request-lab-workspace.jsx';
import { ResourceMonitorTab } from './components/resource-monitor-tab.jsx';
import { TooltipProvider } from './components/ui/tooltip.jsx';
import { Tabs, TabsContent } from './components/ui/tabs.jsx';
import { BrowserOpenURL } from '../wailsjs/runtime/runtime.js';

const GITHUB_URL = 'https://github.com/d0ublecl1ck/RequestProbe';

const workspaceItems = [
  {
    value: 'request-lab',
    label: '字段探针',
  },
  {
    value: 'resource-monitor',
    label: '资源监听',
  },
];

export default function App() {
  const [activeWorkspaceTab, setActiveWorkspaceTab] = useState('request-lab');

  const openGithub = async () => {
    try {
      await BrowserOpenURL(GITHUB_URL);
    } catch {
      window.open(GITHUB_URL, '_blank', 'noopener,noreferrer');
    }
  };

  return (
    <TooltipProvider>
      <div className="app-shell">
        <Toaster richColors position="top-right" />
        <div className="relative z-10 mx-auto flex h-dvh w-[90vw] min-w-0 max-w-[1728px] overflow-hidden py-[clamp(12px,2vh,32px)]">
          <Tabs
            value={activeWorkspaceTab}
            onValueChange={setActiveWorkspaceTab}
            className="flex h-full min-h-0 w-full min-w-0 flex-col gap-4 overflow-hidden xl:flex-row xl:items-stretch xl:gap-6"
          >
            <WorkspaceSidebar
              title="RequestProbe"
              items={workspaceItems}
              value={activeWorkspaceTab}
              onChange={setActiveWorkspaceTab}
              actionIcon={Github}
              actionLabel="打开 GitHub"
              onAction={openGithub}
            />

            <MainContainer>
              <TabsContent value="request-lab" className="mt-0 h-full overflow-hidden">
                <RequestLabWorkspace />
              </TabsContent>

              <TabsContent value="resource-monitor" className="mt-0 h-full overflow-hidden">
                <ResourceMonitorTab />
              </TabsContent>
            </MainContainer>
          </Tabs>
        </div>
      </div>
    </TooltipProvider>
  );
}
