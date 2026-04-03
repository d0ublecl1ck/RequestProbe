import React from 'react';

import { Button } from '../ui/button.jsx';
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip.jsx';
import { SidebarItemGroup } from './sidebar-item-group.jsx';

export function WorkspaceSidebar({
  title,
  items,
  value,
  onChange,
  actionIcon: ActionIcon,
  actionLabel,
  onAction,
}) {
  return (
    <aside className="workspace-sidebar glass-panel flex w-full shrink-0 flex-col rounded-[28px] p-4 xl:h-full xl:w-[272px]">
      <div className="flex items-center justify-between border-b border-border/50 px-3 pb-4 pt-2">
        <p className="text-[13px] font-medium uppercase tracking-[0.24em] text-slate-700">{title}</p>
        {ActionIcon && onAction ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                aria-label={actionLabel}
                onClick={onAction}
              >
                <ActionIcon className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{actionLabel}</TooltipContent>
          </Tooltip>
        ) : null}
      </div>

      <SidebarItemGroup items={items} value={value} onChange={onChange} />
    </aside>
  );
}
