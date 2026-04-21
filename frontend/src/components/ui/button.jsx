import * as React from 'react';
import { cva } from 'class-variance-authority';

import { cn } from '../../lib/utils';

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap border text-[13px] font-semibold uppercase tracking-[0.18em] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 ring-offset-background',
  {
    variants: {
      variant: {
        default: 'border-primary bg-primary text-primary-foreground hover:border-[var(--brand-red-dark)] hover:bg-[var(--brand-red-dark)]',
        destructive: 'border-destructive bg-destructive text-destructive-foreground hover:bg-destructive/90',
        outline: 'border-foreground bg-background text-foreground hover:bg-foreground hover:text-background',
        secondary: 'border-border bg-secondary text-secondary-foreground hover:border-foreground hover:bg-white',
        ghost: 'border-transparent bg-transparent text-foreground hover:border-border hover:bg-accent',
        link: 'text-primary underline-offset-4 hover:underline',
        contrast: 'border-foreground bg-foreground text-background hover:bg-foreground/90'
      },
      size: {
        default: 'h-11 px-4 py-2',
        sm: 'h-9 px-3 text-[12px]',
        lg: 'h-12 px-8',
        icon: 'h-11 w-11'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'default'
    }
  }
);

const Button = React.forwardRef(({ className, variant, size, ...props }, ref) => (
  <button
    className={cn(buttonVariants({ variant, size, className }))}
    ref={ref}
    {...props}
  />
));
Button.displayName = 'Button';

export { Button, buttonVariants };
