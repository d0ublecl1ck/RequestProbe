import React from 'react';

export function MainContainer({ children }) {
  return <div className="min-h-0 min-w-0 flex-1 overflow-hidden">{children}</div>;
}
