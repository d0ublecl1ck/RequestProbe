import React, { useEffect, useState } from 'react';
import { FolderOpen, RotateCcw, Save, Settings2 } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from './ui/button.jsx';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card.jsx';
import { Input } from './ui/input.jsx';
import {
  ChooseResourceMonitorSaveRoot,
  GetResourceMonitorSettings,
  ResetResourceMonitorSaveRoot,
  UpdateResourceMonitorSaveRoot,
} from '../../wailsjs/go/main/App.js';

export function SettingsTab() {
  const [saveRootDir, setSaveRootDir] = useState('');
  const [defaultSaveRootDir, setDefaultSaveRootDir] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isBrowsing, setIsBrowsing] = useState(false);
  const [isResetting, setIsResetting] = useState(false);

  useEffect(() => {
    let mounted = true;

    const loadSettings = async () => {
      try {
        const settings = await GetResourceMonitorSettings();
        if (!mounted || !settings) return;
        setSaveRootDir(settings.saveRootDir || '');
        setDefaultSaveRootDir(settings.defaultSaveRootDir || '');
      } catch (error) {
        const message = error?.message || error?.toString() || '加载设置失败';
        toast.error(message);
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    loadSettings();

    return () => {
      mounted = false;
    };
  }, []);

  const saveSettings = async () => {
    setIsSaving(true);
    try {
      const settings = await UpdateResourceMonitorSaveRoot(saveRootDir);
      setSaveRootDir(settings?.saveRootDir || '');
      setDefaultSaveRootDir(settings?.defaultSaveRootDir || '');
      toast.success('保存位置已更新');
    } catch (error) {
      const message = error?.message || error?.toString() || '保存失败';
      toast.error(message);
    } finally {
      setIsSaving(false);
    }
  };

  const browseDirectory = async () => {
    setIsBrowsing(true);
    try {
      const selectedDir = await ChooseResourceMonitorSaveRoot();
      if (selectedDir) {
        setSaveRootDir(selectedDir);
      }
    } catch (error) {
      const message = error?.message || error?.toString() || '选择目录失败';
      toast.error(message);
    } finally {
      setIsBrowsing(false);
    }
  };

  const resetSettings = async () => {
    setIsResetting(true);
    try {
      const settings = await ResetResourceMonitorSaveRoot();
      setSaveRootDir(settings?.saveRootDir || '');
      setDefaultSaveRootDir(settings?.defaultSaveRootDir || '');
      toast.success('已恢复默认保存位置');
    } catch (error) {
      const message = error?.message || error?.toString() || '恢复默认值失败';
      toast.error(message);
    } finally {
      setIsResetting(false);
    }
  };

  return (
    <div className="grid h-full min-h-0 min-w-0 gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1fr)] xl:gap-6">
      <Card className="glass-panel overflow-hidden">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings2 className="h-5 w-5 text-sky-500" />
            应用设置
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <p className="text-sm font-medium text-slate-800">资源监听页面保存文件的位置</p>
            <p className="text-sm text-muted-foreground">
              新的资源监听任务会在这里创建按任务 UUID 隔离的子目录。
            </p>
          </div>

          <div className="space-y-3">
            <Input
              value={saveRootDir}
              onChange={(event) => setSaveRootDir(event.target.value)}
              placeholder={defaultSaveRootDir || '请选择保存目录'}
              className="bg-white/80"
              disabled={isLoading}
            />

            <div className="rounded-xl border border-border/60 bg-white/70 px-4 py-3 text-sm text-slate-600">
              <p className="font-medium text-slate-800">默认目录</p>
              <p className="mt-1 break-all [overflow-wrap:anywhere]">{defaultSaveRootDir || '--'}</p>
            </div>
          </div>

          <div className="flex flex-wrap gap-3">
            <Button
              variant="outline"
              className="gap-2"
              onClick={browseDirectory}
              disabled={isLoading || isBrowsing}
            >
              <FolderOpen className="h-4 w-4" />
              {isBrowsing ? '正在选择...' : '选择目录'}
            </Button>
            <Button
              className="gap-2"
              onClick={saveSettings}
              disabled={isLoading || isSaving || !saveRootDir.trim()}
            >
              <Save className="h-4 w-4" />
              {isSaving ? '正在保存...' : '保存设置'}
            </Button>
            <Button
              variant="secondary"
              className="gap-2"
              onClick={resetSettings}
              disabled={isLoading || isResetting}
            >
              <RotateCcw className="h-4 w-4" />
              {isResetting ? '正在恢复...' : '恢复默认值'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
