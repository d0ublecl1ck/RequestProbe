import React from 'react';

export function WorkspaceSidebar({ items, value, onChange }) {
  return (
    <aside className="workspace-sidebar flex w-full shrink-0 flex-col xl:h-full xl:w-[292px]">
      <div className="workspace-sidebar-brand">
        <span className="workspace-sidebar-badge">RequestProbe</span>
        <div>
          <p className="workspace-sidebar-title">请求工作台</p>
          <p className="workspace-sidebar-copy">
            在一个界面里完成请求分析、资源监听和全局设置。
          </p>
        </div>
      </div>

      <div className="workspace-nav">
        {items.map((item) => {
          const isActive = item.value === value;
          return (
            <button
              key={item.value}
              type="button"
              onClick={() => onChange(item.value)}
              className={`workspace-nav-item ${isActive ? 'workspace-nav-item-active' : ''}`}
            >
              <span className="workspace-nav-item-label">{item.label}</span>
              {item.subtitle ? <span className="workspace-nav-item-copy">{item.subtitle}</span> : null}
            </button>
          );
        })}
      </div>
    </aside>
  );
}
