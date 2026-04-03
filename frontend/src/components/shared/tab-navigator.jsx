import React from 'react';

import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs.jsx';

export function TabNavigator({
  value,
  onValueChange,
  defaultValue,
  tabs,
  className = '',
  listClassName = '',
  contentClassName = 'mt-0',
}) {
  return (
    <Tabs
      value={value}
      onValueChange={onValueChange}
      defaultValue={defaultValue}
      className={className}
    >
      <TabsList className={listClassName}>
        {tabs.map((tab) => (
          <TabsTrigger key={tab.value} value={tab.value} disabled={tab.disabled}>
            {tab.label}
          </TabsTrigger>
        ))}
      </TabsList>
      {tabs.map((tab) => (
        <TabsContent
          key={tab.value}
          value={tab.value}
          className={`${contentClassName} ${tab.contentClassName || ''}`}
        >
          {tab.content}
        </TabsContent>
      ))}
    </Tabs>
  );
}
