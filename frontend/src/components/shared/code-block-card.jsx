import React from 'react';
import { Copy } from 'lucide-react';

import { Button } from '../ui/button.jsx';
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip.jsx';
import { ContainerCard } from './container-card.jsx';
import { EmptyState } from './empty-state.jsx';

export function CodeBlockCard({
  title,
  description,
  code,
  emptyMessage,
  onCopy,
  copyLabel = '复制代码',
  className = '',
  style,
}) {
  const action = onCopy ? (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="gap-2"
          onClick={() => onCopy(code)}
          disabled={!code}
        >
          <Copy className="h-4 w-4" />
          {copyLabel}
        </Button>
      </TooltipTrigger>
      <TooltipContent>{copyLabel}</TooltipContent>
    </Tooltip>
  ) : null;

  return (
    <ContainerCard
      title={title}
      description={description}
      action={action}
      className={className}
      style={style}
    >
      {code ? (
        <pre className="code-block whitespace-pre-wrap">{code}</pre>
      ) : (
        <EmptyState message={emptyMessage} />
      )}
    </ContainerCard>
  );
}
