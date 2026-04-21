import React from 'react';

export function FormSection({ title, description, children, className = '' }) {
  return (
    <section className={`space-y-3 border-t border-border pt-4 first:border-t-0 first:pt-0 ${className}`}>
      <div>
        <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-muted-foreground">{title}</p>
        {description ? <p className="mt-1 text-sm text-muted-foreground">{description}</p> : null}
      </div>
      {children}
    </section>
  );
}
