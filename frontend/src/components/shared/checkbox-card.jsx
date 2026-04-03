import React from 'react';

import { Checkbox } from '../ui/checkbox.jsx';

export function CheckboxCard({
  checked,
  onCheckedChange,
  label,
  className = '',
  labelClassName = '',
}) {
  return (
    <label className={`flex items-center gap-2 rounded-md border border-border/60 bg-white/70 px-3 py-3 text-sm ${className}`}>
      <Checkbox checked={checked} onCheckedChange={onCheckedChange} />
      <span className={labelClassName}>{label}</span>
    </label>
  );
}
