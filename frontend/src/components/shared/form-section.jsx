import React from 'react';

export function FormSection({ title, description, children, className = '' }) {
  return (
    <section className={`space-y-3 ${className}`}>
      <div>
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">{title}</p>
        {description ? <p className="mt-1 text-sm text-muted-foreground">{description}</p> : null}
      </div>
      {children}
    </section>
  );
}
