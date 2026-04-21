import React from 'react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../ui/card.jsx';

export function ContainerCard({
  title,
  description,
  action,
  icon: Icon,
  children,
  className = '',
  headerClassName = '',
  contentClassName = '',
  ...props
}) {
  return (
    <Card className={`glass-panel min-w-0 overflow-hidden ${className}`} {...props}>
      {(title || description || action || Icon) && (
        <CardHeader className={headerClassName}>
          <div className="flex items-start justify-between gap-4">
            <div>
              {title && (
                <CardTitle className="flex items-center gap-2">
                  {Icon ? <Icon className="h-5 w-5 text-[var(--brand-red)]" /> : null}
                  <span className={Icon ? '' : 'section-title'}>{title}</span>
                </CardTitle>
              )}
              {description ? <CardDescription>{description}</CardDescription> : null}
            </div>
            {action}
          </div>
        </CardHeader>
      )}
      <CardContent className={`min-w-0 ${contentClassName}`}>{children}</CardContent>
    </Card>
  );
}
