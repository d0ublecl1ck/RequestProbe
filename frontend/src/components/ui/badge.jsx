import * as React from 'react';
import { cva } from 'class-variance-authority';

import { cn } from '../../lib/utils';

const badgeVariants = cva(
  'inline-flex items-center border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.16em] transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-primary text-primary-foreground',
        secondary: 'border-border bg-secondary text-secondary-foreground',
        destructive: 'border-transparent bg-destructive text-destructive-foreground',
        outline: 'border-border bg-background text-foreground',
        success: 'border-transparent bg-emerald-500/15 text-emerald-700',
        warning: 'border-transparent bg-amber-500/15 text-amber-700',
        info: 'border-transparent bg-sky-500/15 text-sky-700'
      }
    },
    defaultVariants: {
      variant: 'default'
    }
  }
);

function Badge({ className, variant, ...props }) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { Badge, badgeVariants };
