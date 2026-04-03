import React from 'react';

export function EmptyState({ message, className = '' }) {
  return (
    <div className={`rounded-xl border border-dashed border-border bg-white/70 px-6 py-10 text-center text-sm text-muted-foreground ${className}`}>
      {message}
    </div>
  );
}
