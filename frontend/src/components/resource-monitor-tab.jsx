import React, { useEffect, useMemo, useState } from 'react';
import {
  Code2,
  FolderOpen,
  Loader2,
  PauseCircle,
  Play,
  Power,
  RefreshCcw,
  Download,
  Link2,
} from 'lucide-react';
import { toast } from 'sonner';

import { Badge } from './ui/badge.jsx';
import { Button } from './ui/button.jsx';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from './ui/card.jsx';
import { Checkbox } from './ui/checkbox.jsx';
import { Input } from './ui/input.jsx';
import { ScrollArea } from './ui/scroll-area.jsx';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table.jsx';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs.jsx';
import { OpenerSelect } from './resource-monitor/opener-select.jsx';
import { EventsOn } from '../../wailsjs/runtime/runtime.js';
import {
  GetResourceMonitorSettings,
  DownloadSelectedResources,
  EndResourceMonitor,
  GetCommonResourceExtensions,
  GetResourceMonitorTask,
  OpenResourceMonitorDownloadDir,
  PauseResourceMonitor,
  ResumeResourceMonitor,
  StartResourceMonitor,
} from '../../wailsjs/go/main/App.js';

const DEFAULT_EXTENSIONS = ['js', 'wasm'];
const EMPTY_ITEMS = [];
const OPENERS = [
  {
    value: 'finder',
    label: 'Finder',
    subtitle: '打开任务目录',
    icon: FolderOpen,
  },
  {
    value: 'vscode',
    label: 'VS Code',
    subtitle: '打开任务目录',
    icon: Code2,
  },
];

const statusMeta = {
  idle: { label: '未开始', variant: 'outline' },
  running: { label: '监听中', variant: 'success' },
  paused: { label: '已暂停', variant: 'warning' },
  ended: { label: '已结束', variant: 'outline' },
  error: { label: '异常', variant: 'destructive' },
};

const formatBytes = (bytes) => {
  if (!bytes) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / Math.pow(1024, index);
  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
};

const formatTime = (value) => {
  if (!value) return '--';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
};

const getEventPayload = (eventArgs) => {
  if (Array.isArray(eventArgs) && eventArgs.length > 0) {
    return eventArgs[0];
  }
  return eventArgs;
};

const normalizeTask = (value) => {
  if (!value || typeof value !== 'object') {
    return null;
  }

  const resources = Array.isArray(value.resources) ? value.resources : EMPTY_ITEMS;
  const requests = Array.isArray(value.requests) ? value.requests : EMPTY_ITEMS;
  const hasIdentity = Boolean(
    value.taskId
      || value.status
      || value.downloadDir
      || value.createdAt
      || value.updatedAt
      || resources.length
      || requests.length,
  );

  if (!hasIdentity) {
    return null;
  }

  return {
    ...value,
    resources,
    requests,
  };
};

export function ResourceMonitorTab() {
  const [url, setUrl] = useState('');
  const [availableExtensions, setAvailableExtensions] = useState(DEFAULT_EXTENSIONS);
  const [selectedExtensions, setSelectedExtensions] = useState(DEFAULT_EXTENSIONS);
  const [task, setTask] = useState(null);
  const [selectedIds, setSelectedIds] = useState({});
  const [isStarting, setIsStarting] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);
  const [isMutatingTask, setIsMutatingTask] = useState(false);
  const [isOpeningFinder, setIsOpeningFinder] = useState(false);
  const [isOpeningVSCode, setIsOpeningVSCode] = useState(false);
  const [selectedOpener, setSelectedOpener] = useState('finder');
  const [saveRootDir, setSaveRootDir] = useState('');
  const [activeView, setActiveView] = useState('resources');

  const resources = task?.resources ?? EMPTY_ITEMS;
  const requests = task?.requests ?? EMPTY_ITEMS;
  const taskStatus = task?.status || 'idle';
  const statusInfo = statusMeta[taskStatus] || statusMeta.idle;
  const activeSelectionIds = useMemo(
    () => Object.entries(selectedIds).filter(([, checked]) => checked).map(([id]) => id),
    [selectedIds],
  );
  const allChecked = resources.length > 0 && activeSelectionIds.length === resources.length;
  const someChecked = activeSelectionIds.length > 0 && activeSelectionIds.length < resources.length;

  useEffect(() => {
    let mounted = true;

    const loadInitialState = async () => {
      try {
        const [extensions, currentTask, settings] = await Promise.all([
          GetCommonResourceExtensions(),
          GetResourceMonitorTask(),
          GetResourceMonitorSettings(),
        ]);

        if (!mounted) return;

        const nextExtensions = extensions?.length ? extensions : DEFAULT_EXTENSIONS;
        setAvailableExtensions(nextExtensions);
        setSaveRootDir(settings?.saveRootDir || '');

        const normalizedTask = normalizeTask(currentTask);
        if (normalizedTask) {
          setTask(normalizedTask);
          if (normalizedTask.url) {
            setUrl(normalizedTask.url);
          }
          if (normalizedTask.selectedExtensions?.length) {
            setSelectedExtensions(normalizedTask.selectedExtensions);
          }
        } else {
          setSelectedExtensions((prev) => prev.length ? prev : DEFAULT_EXTENSIONS.filter((item) => nextExtensions.includes(item)));
        }
      } catch (error) {
        console.error('初始化资源监听失败:', error);
        toast.error('初始化资源监听失败');
      }
    };

    loadInitialState();

    const unsubscribe = EventsOn('resource-monitor-event', (...args) => {
      const payload = getEventPayload(args);
      if (!payload) return;
      if (payload.task) {
        setTask(normalizeTask(payload.task));
      }
      if (payload.type === 'resources_downloaded' && payload.download) {
        toast.success(`已下载 ${payload.download.downloadedIds?.length || 0} 个资源`);
      }
      if (payload.type === 'worker_log' && payload.message) {
        console.warn('[resource-monitor-worker]', payload.message);
      }
    });

    return () => {
      mounted = false;
      unsubscribe?.();
    };
  }, []);

  useEffect(() => {
    setSelectedIds((prev) => {
      if (resources.length === 0) {
        if (Object.keys(prev).length === 0) {
          return prev;
        }
        return {};
      }

      const next = {};
      for (const item of resources) {
        if (prev[item.id]) {
          next[item.id] = true;
        }
      }

      const prevKeys = Object.keys(prev);
      const nextKeys = Object.keys(next);
      if (
        prevKeys.length === nextKeys.length
        && prevKeys.every((key) => next[key] === prev[key])
      ) {
        return prev;
      }
      return next;
    });
  }, [resources]);

  const toggleExtension = (extension, checked) => {
    setSelectedExtensions((prev) => {
      const next = new Set(prev);
      if (checked) {
        next.add(extension);
      } else {
        next.delete(extension);
      }
      return Array.from(next).sort();
    });
  };

  const startTask = async () => {
    if (selectedExtensions.length === 0) {
      toast.warning('请至少勾选一个文件后缀');
      return;
    }

    setIsStarting(true);
    try {
      const nextTask = await StartResourceMonitor(url.trim(), selectedExtensions);
      setTask(normalizeTask(nextTask));
      setSelectedIds({});
      toast.success('资源监听已启动');
    } catch (error) {
      const message = error?.message || error?.toString() || '启动失败';
      toast.error(message);
    } finally {
      setIsStarting(false);
    }
  };

  const runTaskMutation = async (handler, successMessage) => {
    setIsMutatingTask(true);
    try {
      const nextTask = await handler();
      setTask(normalizeTask(nextTask));
      toast.success(successMessage);
    } catch (error) {
      const message = error?.message || error?.toString() || '操作失败';
      toast.error(message);
    } finally {
      setIsMutatingTask(false);
    }
  };

  const downloadSelected = async () => {
    if (activeSelectionIds.length === 0) {
      toast.warning('请先勾选要下载的资源');
      return;
    }

    setIsDownloading(true);
    try {
      const result = await DownloadSelectedResources(activeSelectionIds);
      if (result?.downloadedIds?.length) {
        toast.success(`已下载 ${result.downloadedIds.length} 个资源`);
      } else {
        toast.warning('所选资源均已存在或不可下载');
      }
      const refreshedTask = await GetResourceMonitorTask();
      setTask(normalizeTask(refreshedTask));
    } catch (error) {
      const message = error?.message || error?.toString() || '下载失败';
      toast.error(message);
    } finally {
      setIsDownloading(false);
    }
  };

  const openDir = async (opener) => {
    const setLoading = opener === 'finder' ? setIsOpeningFinder : setIsOpeningVSCode;
    setLoading(true);
    try {
      await OpenResourceMonitorDownloadDir(opener);
    } catch (error) {
      const message = error?.message || error?.toString() || '打开目录失败';
      toast.error(message);
    } finally {
      setLoading(false);
    }
  };

  const toggleAll = (checked) => {
    if (!checked) {
      setSelectedIds({});
      return;
    }
    const next = {};
    for (const item of resources) {
      next[item.id] = true;
    }
    setSelectedIds(next);
  };

  const heroMetrics = [
    {
      label: '任务状态',
      value: statusInfo.label,
      copy: task?.taskId ? `任务 ${task.taskId.slice(0, 8)} 已建立上下文` : '等待创建新的监听任务',
    },
    {
      label: '资源命中',
      value: `${resources.length}`,
      copy: activeSelectionIds.length > 0 ? `当前已选中 ${activeSelectionIds.length} 个资源` : '资源列表会按内容哈希去重',
    },
    {
      label: '请求流',
      value: `${requests.length}`,
      copy: taskStatus === 'running' ? '当前会话正在持续接收新请求' : '启动任务后同步采集页面请求',
    },
  ];

  return (
    <div className="workspace-page">
      <section className="workspace-section">
        <div className="workspace-section-header">
          <div>
            <h2 className="workspace-section-title">任务控制与结果浏览</h2>
            <p className="workspace-section-copy">
              左侧配置目标、后缀和任务状态，右侧浏览命中资源与请求流。
            </p>
          </div>
          <Badge variant={statusInfo.variant}>{statusInfo.label}</Badge>
        </div>

        <div className="editorial-grid editorial-grid-2 p-3 xl:p-4">
          <div className="flex min-w-0 flex-col gap-4 xl:min-h-0 xl:gap-5 xl:overflow-y-auto xl:pr-1">
            <Card className="glass-panel overflow-hidden">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Link2 className="h-5 w-5 text-emerald-500" />
              任务启动
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-2">
              <p className="monitor-label">目标 URL（可选）</p>
              <Input
                value={url}
                onChange={(event) => setUrl(event.target.value)}
                placeholder="https://example.com/app"
                className="bg-white"
              />
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <p className="monitor-label">监听后缀</p>
                <Badge variant="outline">{selectedExtensions.length} 项</Badge>
              </div>
              <div className="grid grid-cols-2 gap-2">
                {availableExtensions.map((extension) => (
                  <label key={extension} className="monitor-option cursor-pointer">
                    <span className="font-mono text-[13px]">.{extension}</span>
                    <Checkbox
                      checked={selectedExtensions.includes(extension)}
                      onCheckedChange={(checked) => toggleExtension(extension, Boolean(checked))}
                    />
                  </label>
                ))}
              </div>
            </div>

            <Button
              className="w-full gap-2"
              onClick={startTask}
              disabled={isStarting || taskStatus === 'running' || taskStatus === 'paused'}
            >
              {isStarting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
              {isStarting ? '正在启动...' : '开始监听'}
            </Button>
          </CardContent>
            </Card>

            <Card className="glass-panel overflow-hidden monitor-status-card">
          <CardHeader>
            <CardTitle className="flex items-center justify-between gap-3">
              <span>任务状态</span>
              <Badge variant={statusInfo.variant}>{statusInfo.label}</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-3 sm:grid-cols-2">
              <div className="monitor-metric">
                <p className="monitor-label">任务 UUID</p>
                <p className="monitor-mono">{task?.taskId || '--'}</p>
              </div>
              <div className="monitor-metric">
                <p className="monitor-label">已命中资源</p>
                <p className="monitor-mono">{resources.length}</p>
              </div>
              <div className="monitor-metric sm:col-span-2">
                <p className="monitor-label">已监听请求</p>
                <p className="monitor-mono">{requests.length}</p>
              </div>
            </div>

            <div className="monitor-metric">
              <p className="monitor-label">下载目录</p>
              <p className="break-all text-sm text-slate-200/88">{task?.downloadDir || '--'}</p>
            </div>

            {!task?.downloadDir && saveRootDir && (
              <div className="monitor-metric">
                <p className="monitor-label">默认保存根目录</p>
                <p className="break-all text-sm text-slate-200/88 [overflow-wrap:anywhere]">
                  {saveRootDir}
                </p>
              </div>
            )}

            {task?.lastError && (
              <div className="rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-100">
                {task.lastError}
              </div>
            )}

            <div className="grid gap-2 sm:grid-cols-3">
              <Button
                variant="secondary"
                className="gap-2"
                onClick={() => runTaskMutation(PauseResourceMonitor, '已停止监听')}
                disabled={isMutatingTask || taskStatus !== 'running'}
              >
                {isMutatingTask && taskStatus === 'running' ? <Loader2 className="h-4 w-4 animate-spin" /> : <PauseCircle className="h-4 w-4" />}
                停止监听
              </Button>
              <Button
                variant="outline"
                className="gap-2"
                onClick={() => runTaskMutation(ResumeResourceMonitor, '已继续监听')}
                disabled={isMutatingTask || taskStatus !== 'paused'}
              >
                <RefreshCcw className="h-4 w-4" />
                继续监听
              </Button>
              <Button
                variant="destructive"
                className="gap-2"
                onClick={() => runTaskMutation(EndResourceMonitor, '任务已结束')}
                disabled={isMutatingTask || !task || taskStatus === 'ended' || taskStatus === 'idle'}
              >
                <Power className="h-4 w-4" />
                结束任务
              </Button>
            </div>

            <OpenerSelect
              options={OPENERS}
              selectedValue={selectedOpener}
              onSelect={setSelectedOpener}
              onOpen={openDir}
              disabled={!task?.downloadDir}
              loadingValues={{
                finder: isOpeningFinder,
                vscode: isOpeningVSCode,
              }}
            />
          </CardContent>
            </Card>
          </div>

          <Card className="glass-panel flex min-h-[420px] min-w-0 flex-col overflow-hidden xl:min-h-0">
            <Tabs
              value={activeView}
              onValueChange={setActiveView}
              className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden"
            >
              <CardHeader className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                <div className="flex flex-col gap-2">
                  <div className="flex items-center gap-3">
                    <CardTitle>{activeView === 'resources' ? '命中资源列表' : '请求监听列表'}</CardTitle>
                    <TabsList className="bg-[var(--paper-alt)]">
                      <TabsTrigger value="resources">资源</TabsTrigger>
                      <TabsTrigger value="requests">请求</TabsTrigger>
                    </TabsList>
                  </div>
                  <div className="grid gap-2 sm:grid-cols-3">
                    {heroMetrics.map((item) => (
                      <div key={item.label} className="metric-card">
                        <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">{item.label}</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">{item.value}</p>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  {activeView === 'resources' && (
                    <Button
                      variant="outline"
                      className="gap-2"
                      onClick={downloadSelected}
                      disabled={isDownloading || activeSelectionIds.length === 0}
                    >
                      {isDownloading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
                      下载选中项
                    </Button>
                  )}
                </div>
              </CardHeader>

              <CardContent className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden p-0">
                <TabsContent value="resources" className="mt-0 flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
              {resources.length > 0 ? (
                <ScrollArea className="min-h-0 flex-1">
                  <div className="p-4">
                    <div className="mb-3 flex items-center justify-between border border-border bg-[var(--paper-alt)] px-4 py-3">
                      <label className="flex cursor-pointer items-center gap-3 text-sm font-medium text-slate-700">
                        <Checkbox
                          checked={allChecked}
                          onCheckedChange={(checked) => toggleAll(Boolean(checked))}
                        />
                        <span>{allChecked ? '已全选' : someChecked ? '部分已选' : '全选资源'}</span>
                      </label>
                      <Badge variant="info">{activeSelectionIds.length} / {resources.length} 已选</Badge>
                    </div>

                    <div className="overflow-hidden border border-border bg-white">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead className="w-[68px]">选择</TableHead>
                            <TableHead className="w-[110px]">类型</TableHead>
                            <TableHead>资源地址</TableHead>
                            <TableHead className="w-[120px]">大小</TableHead>
                            <TableHead className="w-[140px]">状态</TableHead>
                            <TableHead className="w-[190px]">文件名</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {resources.map((resource) => (
                            <TableRow key={resource.id}>
                              <TableCell>
                                <Checkbox
                                  checked={Boolean(selectedIds[resource.id])}
                                  onCheckedChange={(checked) =>
                                    setSelectedIds((prev) => ({
                                      ...prev,
                                      [resource.id]: Boolean(checked),
                                    }))
                                  }
                                />
                              </TableCell>
                              <TableCell>
                                <div className="flex items-center gap-2">
                                  <Badge variant="outline" className="font-mono">.{resource.extension}</Badge>
                                  {resource.downloaded && <Badge variant="success">已下载</Badge>}
                                </div>
                              </TableCell>
                              <TableCell className="align-top">
                                <div className="space-y-1">
                                  <p className="break-all text-sm font-medium text-slate-800 [overflow-wrap:anywhere]">{resource.url}</p>
                                  <p className="monitor-mono text-[11px] text-slate-500">Hash {resource.hash.slice(0, 12)}</p>
                                  <p className="text-xs text-muted-foreground">首次命中 {formatTime(resource.firstSeenAt)}</p>
                                </div>
                              </TableCell>
                              <TableCell className="font-medium text-slate-700">{formatBytes(resource.size)}</TableCell>
                              <TableCell>
                                <div className="space-y-1">
                                  <Badge variant={resource.statusCode >= 400 ? 'destructive' : 'secondary'}>
                                    {resource.statusCode || '--'}
                                  </Badge>
                                  {resource.mimeType && (
                                    <p className="break-all text-xs text-muted-foreground [overflow-wrap:anywhere]">{resource.mimeType}</p>
                                  )}
                                </div>
                              </TableCell>
                              <TableCell className="align-top">
                                <div className="space-y-1">
                                  <p className="break-all text-sm font-medium text-slate-700 [overflow-wrap:anywhere]">{resource.suggestedFileName}</p>
                                  {resource.downloadedPath && (
                                    <p className="break-all text-xs text-muted-foreground [overflow-wrap:anywhere]">{resource.downloadedPath}</p>
                                  )}
                                </div>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  </div>
                </ScrollArea>
              ) : (
                <div className="flex min-h-0 flex-1 items-center justify-center px-8 text-center text-sm text-muted-foreground">
                  当前还没有命中资源。启动监听后，这里会实时出现去重后的资源记录。
                </div>
              )}
            </TabsContent>

            <TabsContent value="requests" className="mt-0 flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
              {requests.length > 0 ? (
                <ScrollArea className="min-h-0 flex-1">
                  <div className="p-4">
                    <div className="overflow-hidden border border-border bg-white">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead className="w-[96px]">方法</TableHead>
                            <TableHead className="w-[148px]">状态</TableHead>
                            <TableHead className="w-[120px]">类型</TableHead>
                            <TableHead>请求地址</TableHead>
                            <TableHead className="w-[260px]">摘要</TableHead>
                            <TableHead className="w-[180px]">时间</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {requests.map((request) => (
                            <TableRow key={request.id}>
                              <TableCell>
                                <Badge variant="outline" className="font-mono">{request.method || '--'}</Badge>
                              </TableCell>
                              <TableCell className="align-top">
                                <div className="space-y-1">
                                  <Badge variant={request.failed || request.statusCode >= 400 ? 'destructive' : 'secondary'}>
                                    {request.failed ? `失败 ${request.statusCode || ''}`.trim() : (request.statusCode || '--')}
                                  </Badge>
                                  {request.failureText && (
                                    <p className="break-all text-xs text-red-500 [overflow-wrap:anywhere]">{request.failureText}</p>
                                  )}
                                </div>
                              </TableCell>
                              <TableCell className="align-top">
                                <div className="space-y-1">
                                  <Badge variant="outline">{request.resourceType || '--'}</Badge>
                                  {request.mimeType && (
                                    <p className="break-all text-xs text-muted-foreground [overflow-wrap:anywhere]">{request.mimeType}</p>
                                  )}
                                </div>
                              </TableCell>
                              <TableCell className="align-top">
                                <p className="break-all text-sm font-medium text-slate-800 [overflow-wrap:anywhere]">{request.url}</p>
                              </TableCell>
                              <TableCell className="align-top">
                                <div className="space-y-1 text-xs text-muted-foreground">
                                  {request.requestBodyPreview ? (
                                    <p className="break-all [overflow-wrap:anywhere]">请求: {request.requestBodyPreview}</p>
                                  ) : (
                                    <p>请求体为空</p>
                                  )}
                                  {request.responseBodyPreview && (
                                    <p className="break-all [overflow-wrap:anywhere]">响应: {request.responseBodyPreview}</p>
                                  )}
                                </div>
                              </TableCell>
                              <TableCell className="text-xs text-muted-foreground">
                                {formatTime(request.firstSeenAt)}
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  </div>
                </ScrollArea>
              ) : (
                <div className="flex min-h-0 flex-1 items-center justify-center px-8 text-center text-sm text-muted-foreground">
                  当前还没有请求事件。任务启动并加载页面后，这里会实时展示请求流。
                </div>
              )}
            </TabsContent>
              </CardContent>
            </Tabs>
          </Card>
        </div>
      </section>
    </div>
  );
}
