import React from 'react';

export function SidebarItemGroup({ items, value, onChange }) {
  return (
    <div className="mt-4 grid grid-cols-2 gap-2 xl:flex xl:flex-col">
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
  );
}
