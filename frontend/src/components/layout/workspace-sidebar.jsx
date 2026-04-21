import React from 'react';

export function WorkspaceSidebar({ items, value, onChange, currentTitle, currentDescription }) {
  return (
    <aside className="workspace-sidebar flex w-full shrink-0 flex-col xl:h-full xl:w-[292px]">
      <div className="workspace-sidebar-brand">
        <span className="workspace-sidebar-badge">RequestProbe</span>
        <div>
          <p className="workspace-sidebar-title">{currentTitle}</p>
          {currentDescription ? <p className="workspace-sidebar-copy">{currentDescription}</p> : null}
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
            </button>
          );
        })}
      </div>
    </aside>
  );
}
