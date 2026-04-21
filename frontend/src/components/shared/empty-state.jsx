import React from 'react';

export function EmptyState({ message, className = '' }) {
  return (
    <div className={`border border-dashed border-border bg-[var(--paper-alt)] px-6 py-10 text-center text-sm leading-7 text-muted-foreground ${className}`}>
      {message}
    </div>
  );
}
