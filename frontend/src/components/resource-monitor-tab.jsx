import React, { useEffect, useMemo, useState } from 'react';
import {
  Code2,
  ChevronDown,
  FolderOpen,
  Loader2,
  PauseCircle,
  Play,
  Power,
  RefreshCcw,
  Download,
  ExternalLink,
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
import { EventsOn } from '../../wailsjs/runtime/runtime.js';
import {
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
  const [isOpenerMenuOpen, setIsOpenerMenuOpen] = useState(false);

  const resources = task?.resources || [];
  const taskStatus = task?.status || 'idle';
  const statusInfo = statusMeta[taskStatus] || statusMeta.idle;
  const currentOpener = OPENERS.find((item) => item.value === selectedOpener) || OPENERS[0];
  const CurrentOpenerIcon = currentOpener.icon;
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
        const [extensions, currentTask] = await Promise.all([
          GetCommonResourceExtensions(),
          GetResourceMonitorTask(),
        ]);

        if (!mounted) return;

        const nextExtensions = extensions?.length ? extensions : DEFAULT_EXTENSIONS;
        setAvailableExtensions(nextExtensions);

        if (currentTask) {
          setTask(currentTask);
          if (currentTask.url) {
            setUrl(currentTask.url);
          }
          if (currentTask.selectedExtensions?.length) {
            setSelectedExtensions(currentTask.selectedExtensions);
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
        setTask(payload.task);
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
    if (!isOpenerMenuOpen) {
      return undefined;
    }

    const handlePointerDown = (event) => {
      const menuRoot = document.querySelector('[data-opener-select-root="true"]');
      if (menuRoot && !menuRoot.contains(event.target)) {
        setIsOpenerMenuOpen(false);
      }
    };

    window.addEventListener('pointerdown', handlePointerDown);
    return () => window.removeEventListener('pointerdown', handlePointerDown);
  }, [isOpenerMenuOpen]);

  useEffect(() => {
    setSelectedIds((prev) => {
      const next = {};
      for (const item of resources) {
        if (prev[item.id]) {
          next[item.id] = true;
        }
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
      setTask(nextTask);
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
      setTask(nextTask);
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
      setTask(refreshedTask);
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

  return (
    <div className="grid min-w-0 gap-6 xl:grid-cols-[380px_minmax(0,1fr)]">
      <div className="flex min-w-0 flex-col gap-6">
        <Card className="glass-panel overflow-hidden">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Link2 className="h-5 w-5 text-emerald-500" />
              监听任务
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-2">
              <p className="monitor-label">目标 URL（可选）</p>
              <Input
                value={url}
                onChange={(event) => setUrl(event.target.value)}
                placeholder="https://example.com/app"
                className="bg-white/80"
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
            </div>

            <div className="monitor-metric">
              <p className="monitor-label">下载目录</p>
              <p className="break-all text-sm text-slate-200/88">{task?.downloadDir || '--'}</p>
            </div>

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

            <div className="relative" data-opener-select-root="true">
              <div className="opener-select-shell">
                <button
                  type="button"
                  className="opener-select-main"
                  onClick={() => openDir(currentOpener.value)}
                  disabled={!task?.downloadDir || isOpeningFinder || isOpeningVSCode}
                >
                  <span className="opener-select-icon">
                    {(currentOpener.value === 'finder' && isOpeningFinder) || (currentOpener.value === 'vscode' && isOpeningVSCode) ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <CurrentOpenerIcon className="h-4 w-4" />
                    )}
                  </span>
                  <span className="opener-select-label">{currentOpener.label}</span>
                </button>
                <button
                  type="button"
                  className="opener-select-toggle"
                  onClick={() => setIsOpenerMenuOpen((prev) => !prev)}
                  disabled={!task?.downloadDir}
                  aria-haspopup="menu"
                  aria-expanded={isOpenerMenuOpen}
                  aria-label="选择打开方式"
                >
                  <ChevronDown className={`h-4 w-4 transition-transform ${isOpenerMenuOpen ? 'rotate-180' : ''}`} />
                </button>
              </div>

              {isOpenerMenuOpen && (
                <div className="opener-select-menu" role="menu">
                  {OPENERS.map((opener) => {
                    const OpenerIcon = opener.icon;
                    const isActive = opener.value === currentOpener.value;
                    return (
                      <button
                        key={opener.value}
                        type="button"
                        role="menuitem"
                        className={`opener-select-option ${isActive ? 'opener-select-option-active' : ''}`}
                        onClick={() => {
                          setSelectedOpener(opener.value);
                          setIsOpenerMenuOpen(false);
                        }}
                      >
                        <span className="opener-select-option-icon">
                          <OpenerIcon className="h-4 w-4" />
                        </span>
                        <span className="opener-select-option-text">
                          <span className="opener-select-option-title">{opener.label}</span>
                          <span className="opener-select-option-subtitle">{opener.subtitle}</span>
                        </span>
                        {isActive && <ExternalLink className="opener-select-option-mark" />}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card className="glass-panel min-w-0 overflow-hidden">
        <CardHeader className="flex flex-col gap-3 border-b border-border/60 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <CardTitle>命中资源列表</CardTitle>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              className="gap-2"
              onClick={downloadSelected}
              disabled={isDownloading || activeSelectionIds.length === 0}
            >
              {isDownloading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
              下载选中项
            </Button>
          </div>
        </CardHeader>
        <CardContent className="min-w-0 p-0">
          {resources.length > 0 ? (
            <ScrollArea className="h-[calc(100vh-320px)]">
              <div className="p-4">
                <div className="mb-3 flex items-center justify-between rounded-xl border border-border/60 bg-white/70 px-4 py-3">
                  <label className="flex cursor-pointer items-center gap-3 text-sm font-medium text-slate-700">
                    <Checkbox
                      checked={allChecked}
                      onCheckedChange={(checked) => toggleAll(Boolean(checked))}
                    />
                    <span>{allChecked ? '已全选' : someChecked ? '部分已选' : '全选资源'}</span>
                  </label>
                  <Badge variant="info">{activeSelectionIds.length} / {resources.length} 已选</Badge>
                </div>

                <div className="rounded-xl border border-border/60 bg-white/70">
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
            <div className="flex h-[calc(100vh-320px)] items-center justify-center px-8 text-center text-sm text-muted-foreground">
              暂无资源
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
