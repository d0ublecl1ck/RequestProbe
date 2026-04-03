import React from 'react';

export function MetricCard({ label, children, className = '' }) {
  return (
    <div className={`metric-card ${className}`}>
      <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">{label}</p>
      <div className="mt-2">{children}</div>
    </div>
  );
}
