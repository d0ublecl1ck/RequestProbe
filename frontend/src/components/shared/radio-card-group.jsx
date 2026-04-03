import React from 'react';

import { RadioGroup, RadioGroupItem } from '../ui/radio-group.jsx';

export function RadioCardGroup({
  value,
  onValueChange,
  options,
  className = '',
  optionClassName = '',
}) {
  return (
    <RadioGroup value={value} onValueChange={onValueChange} className={className}>
      {options.map((option) => (
        <label
          key={option.value}
          className={`flex items-center justify-between rounded-lg border border-border/60 bg-white/70 px-3 py-2 text-xs font-medium transition hover:border-foreground/30 ${optionClassName}`}
        >
          <span>{option.label}</span>
          <RadioGroupItem value={option.value} />
        </label>
      ))}
    </RadioGroup>
  );
}
