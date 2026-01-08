import React, { useEffect, useMemo, useState } from 'react';
import {
  Copy,
  Loader2,
  Play,
  RefreshCcw,
  ScanSearch,
  Trash2,
  Wand2,
} from 'lucide-react';
import { toast, Toaster } from 'sonner';

import { Button } from './components/ui/button.jsx';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from './components/ui/card.jsx';
import { Input } from './components/ui/input.jsx';
import { Textarea } from './components/ui/textarea.jsx';
import { Badge } from './components/ui/badge.jsx';
import { RadioGroup, RadioGroupItem } from './components/ui/radio-group.jsx';
import { Checkbox } from './components/ui/checkbox.jsx';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './components/ui/tabs.jsx';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './components/ui/table.jsx';
import { ScrollArea } from './components/ui/scroll-area.jsx';
import { Separator } from './components/ui/separator.jsx';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from './components/ui/tooltip.jsx';

import {
  DetectInputType,
  GeneratePythonCode,
  GetDefaultValidationConfig,
  GetRequestSummary,
  GetSupportedEncodings,
  GetTestStatistics,
  ParseRequest,
  ParseRequestWithType,
  TestFieldNecessity,
  TestRequestOnly,
} from '../wailsjs/go/main/App.js';

const defaultValidationConfig = {
  expression: '',
  timeout: 30,
  maxRetries: 3,
  followRedirect: true,
  userAgent: 'RequestProbe/1.0',
  textMatching: {
    enabled: true,
    texts: [],
    matchMode: 'all',
    caseSensitive: false,
  },
  lengthRange: {
    enabled: false,
    minLength: 0,
    maxLength: -1,
  },
  useCustomExpr: false,
  encodingConfig: {
    enabled: false,
    calibrationText: '',
    supportedEncodings: ['UTF-8', 'GBK', 'GB2312', 'Big5'],
    detectedEncoding: 'UTF-8',
  },
};

const badgeVariantByStatus = (statusCode) => {
  if (statusCode >= 200 && statusCode < 300) return 'success';
  if (statusCode >= 300 && statusCode < 400) return 'warning';
  if (statusCode >= 400) return 'destructive';
  return 'info';
};

export default function App() {
  const [inputText, setInputText] = useState('');
  const [inputType, setInputType] = useState('auto');
  const [detectedType, setDetectedType] = useState('');
  const [isParsing, setIsParsing] = useState(false);
  const [parsedRequest, setParsedRequest] = useState(null);
  const [pythonCode, setPythonCode] = useState('');
  const [requestSummary, setRequestSummary] = useState({});
  const [validationType, setValidationType] = useState('textMatching');
  const [validationConfig, setValidationConfig] = useState(defaultValidationConfig);
  const [textMatchingInput, setTextMatchingInput] = useState('');
  const [testResult, setTestResult] = useState(null);
  const [testStatistics, setTestStatistics] = useState({});
  const [isTestRunning, setIsTestRunning] = useState(false);
  const [singleTestResult, setSingleTestResult] = useState(null);
  const [isSingleTesting, setIsSingleTesting] = useState(false);

  useEffect(() => {
    let isMounted = true;

    const loadDefaults = async () => {
      try {
        const config = await GetDefaultValidationConfig();
        const encodings = await GetSupportedEncodings();
        const normalizedConfig = {
          ...config,
          timeout: config.timeout / 1000000000,
          encodingConfig: {
            ...config.encodingConfig,
            supportedEncodings: encodings || config.encodingConfig?.supportedEncodings || defaultValidationConfig.encodingConfig.supportedEncodings,
          },
        };
        if (isMounted) {
          setValidationConfig(normalizedConfig);
          setTextMatchingInput((normalizedConfig.textMatching?.texts || []).join('\n'));
        }
      } catch (error) {
        console.error('加载默认配置失败:', error);
      }
    };

    loadDefaults();

    return () => {
      isMounted = false;
    };
  }, []);

  useEffect(() => {
    if (!inputText.trim()) {
      setDetectedType('');
      return;
    }

    const handler = setTimeout(async () => {
      try {
        const type = await DetectInputType(inputText);
        setDetectedType(type);
      } catch (error) {
        console.error('检测输入类型失败:', error);
      }
    }, 350);

    return () => clearTimeout(handler);
  }, [inputText]);

  const testResultSummary = useMemo(() => {
    if (!testResult) return null;
    const requiredHeaders = testResult.headerResults?.filter((r) => r.isRequired).length || 0;
    const requiredCookies = testResult.cookieResults?.filter((r) => r.isRequired).length || 0;
    const totalHeaders = testResult.headerResults?.length || 0;
    const totalCookies = testResult.cookieResults?.length || 0;

    return {
      requiredHeaders,
      requiredCookies,
      totalHeaders,
      totalCookies,
      originalPassed: testResult.originalPassed,
      testDuration: testResult.testDuration,
    };
  }, [testResult]);

  const headerTestResults = useMemo(() => testResult?.headerResults || [], [testResult]);
  const cookieTestResults = useMemo(() => testResult?.cookieResults || [], [testResult]);

  const updateValidationType = (type) => {
    setValidationType(type);
    setValidationConfig((prev) => ({
      ...prev,
      textMatching: {
        ...prev.textMatching,
        enabled: type === 'textMatching',
      },
      lengthRange: {
        ...prev.lengthRange,
        enabled: type === 'lengthRange',
      },
    }));
  };

  const updateTextMatching = (value) => {
    setTextMatchingInput(value);
    const texts = value
      .split('\n')
      .map((text) => text.trim())
      .filter((text) => text.length > 0);

    setValidationConfig((prev) => ({
      ...prev,
      textMatching: {
        ...prev.textMatching,
        texts,
      },
    }));
  };

  const parseRequest = async () => {
    if (!inputText.trim()) {
      toast.warning('请输入HTTP请求或Curl命令');
      return;
    }

    setIsParsing(true);
    try {
      const request =
        inputType === 'auto'
          ? await ParseRequest(inputText)
          : await ParseRequestWithType(inputText, inputType);

      setParsedRequest(request);

      const [code, summary] = await Promise.all([
        GeneratePythonCode(request),
        GetRequestSummary(request),
      ]);

      setPythonCode(code);
      setRequestSummary(summary || {});

      toast.success('请求解析成功');
      await testRequestOnly(request);
    } catch (error) {
      const errorMessage = error?.message || error?.toString() || '未知错误';
      toast.error(`解析失败: ${errorMessage}`);
      console.error('解析请求失败:', error);
    } finally {
      setIsParsing(false);
    }
  };

  const testRequestOnly = async (request = parsedRequest) => {
    if (!request) {
      toast.warning('请先解析请求');
      return;
    }

    setIsSingleTesting(true);
    setSingleTestResult(null);

    try {
      const config = {
        ...validationConfig,
        timeout: validationConfig.timeout * 1000000000,
      };

      const result = await TestRequestOnly(request, config);
      setSingleTestResult(result);
      toast.success('请求测试完成，请检查编码并确认');
    } catch (error) {
      const errorMessage = error?.message || error?.toString() || '未知错误';
      toast.error(`请求测试失败: ${errorMessage}`);
      console.error('请求测试失败:', error);
    } finally {
      setIsSingleTesting(false);
    }
  };

  const testFieldNecessity = async () => {
    if (!parsedRequest) {
      toast.warning('请先解析请求');
      return;
    }

    setIsTestRunning(true);
    try {
      const config = {
        ...validationConfig,
        timeout: validationConfig.timeout * 1000000000,
      };
      const result = await TestFieldNecessity(parsedRequest, config);
      setTestResult(result);
      const stats = await GetTestStatistics(result);
      setTestStatistics(stats || {});
      toast.success('字段分析完成');
    } catch (error) {
      const errorMessage = error?.message || error?.toString() || '未知错误';
      toast.error(`测试失败: ${errorMessage}`);
      console.error('测试失败:', error);
    } finally {
      setIsTestRunning(false);
    }
  };

  const clearResults = () => {
    setInputText('');
    setDetectedType('');
    setParsedRequest(null);
    setPythonCode('');
    setRequestSummary({});
    setTextMatchingInput('');
    setValidationConfig((prev) => ({
      ...prev,
      textMatching: {
        ...prev.textMatching,
        texts: [],
      },
    }));
    setTestResult(null);
    setTestStatistics({});
    setSingleTestResult(null);
    toast.success('所有内容已清空');
  };

  const copyCode = async (code) => {
    if (!code) return;
    try {
      await navigator.clipboard.writeText(code);
      toast.success('代码已复制到剪贴板');
    } catch (error) {
      toast.error('复制失败');
    }
  };

  const formatDuration = (duration) => {
    if (!duration) return '0ms';
    const ms = duration / 1000000;
    if (ms < 1000) return `${ms.toFixed(0)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatBytes = (bytes) => {
    if (!bytes) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  };

  const formatHeaders = (headers) => {
    if (!headers) return [];
    return Object.entries(headers).map(([name, value]) => ({ name, value }));
  };

  return (
    <TooltipProvider>
      <div className="app-shell">
        <Toaster richColors position="top-right" />
        <div className="relative z-10 mx-auto flex min-h-screen w-full max-w-[1440px] flex-col gap-6 px-6 py-8">
          <header className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <p className="text-xs uppercase tracking-[0.3em] text-muted-foreground">RequestProbe</p>
              <h1 className="text-3xl font-semibold text-foreground">请求解析与验证控制台</h1>
              <p className="mt-2 max-w-xl text-sm text-muted-foreground">
                将原始请求转成结构化测试资产，快速定位必需字段并生成可复用的 Python 代码片段。
              </p>
            </div>
            <div className="flex flex-wrap gap-3">
              <Button variant="outline" className="gap-2" onClick={() => copyCode(pythonCode)} disabled={!pythonCode}>
                <Copy className="h-4 w-4" />
                复制Python
              </Button>
              <Button variant="contrast" className="gap-2" onClick={parseRequest} disabled={isParsing}>
                {isParsing ? <Loader2 className="h-4 w-4 animate-spin" /> : <Wand2 className="h-4 w-4" />}
                {isParsing ? '解析中...' : '解析请求'}
              </Button>
            </div>
          </header>

          <div className="grid flex-1 gap-6 lg:grid-cols-[minmax(360px,420px)_minmax(0,1fr)]">
            <Card className="glass-panel flex h-[calc(100vh-240px)] flex-col overflow-hidden">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <ScanSearch className="h-5 w-5 text-muted-foreground" />
                  请求输入
                </CardTitle>
                <CardDescription>支持 Raw HTTP 与 curl 命令，自动识别请求格式。</CardDescription>
              </CardHeader>
              <ScrollArea className="flex-1">
                <CardContent className="flex flex-col gap-6">
                  <div className="space-y-3">
                    <p className="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">输入类型</p>
                    <RadioGroup
                      value={inputType}
                      onValueChange={setInputType}
                      className="grid grid-cols-3 gap-2"
                    >
                      {[
                        { value: 'auto', label: '自动检测' },
                        { value: 'raw', label: 'Raw HTTP' },
                        { value: 'curl', label: 'Curl 命令' },
                      ].map((item) => (
                        <label
                          key={item.value}
                          className="flex items-center justify-between rounded-lg border border-border/60 bg-white/70 px-3 py-2 text-xs font-medium transition hover:border-foreground/30"
                        >
                          <span>{item.label}</span>
                          <RadioGroupItem value={item.value} />
                        </label>
                      ))}
                    </RadioGroup>
                  </div>

                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <p className="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">请求内容</p>
                      {detectedType && (
                        <Badge variant="outline">检测到：{detectedType === 'curl' ? 'Curl' : detectedType === 'raw' ? 'Raw HTTP' : '未知格式'}</Badge>
                      )}
                    </div>
                    <Textarea
                      value={inputText}
                      onChange={(event) => setInputText(event.target.value)}
                      placeholder="请输入HTTP请求或Curl命令..."
                      className="min-h-[220px] resize-none bg-white/80"
                    />
                  </div>

                  <div className="grid gap-3 sm:grid-cols-2">
                    <Button
                      variant="default"
                      className="gap-2"
                      onClick={parseRequest}
                      disabled={isParsing}
                    >
                      {isParsing ? <Loader2 className="h-4 w-4 animate-spin" /> : <Wand2 className="h-4 w-4" />}
                      {isParsing ? '解析中...' : '解析请求'}
                    </Button>
                    <Button
                      variant="outline"
                      className="gap-2"
                      onClick={() => setInputText('')}
                      disabled={!inputText}
                    >
                      <Trash2 className="h-4 w-4" />
                      清空输入
                    </Button>
                  </div>

                  {parsedRequest && (
                    <>
                      <Separator />
                      <div className="space-y-4">
                        <div>
                          <p className="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">验证配置</p>
                          <p className="text-sm text-muted-foreground">配置校验策略、匹配模式与超时参数。</p>
                        </div>

                        <div className="space-y-3">
                          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">验证方式</p>
                          <RadioGroup
                            value={validationType}
                            onValueChange={updateValidationType}
                            className="grid grid-cols-2 gap-2"
                          >
                            {[{ value: 'textMatching', label: '文本匹配' }, { value: 'lengthRange', label: '响应长度' }].map((item) => (
                              <label
                                key={item.value}
                                className="flex items-center justify-between rounded-lg border border-border/60 bg-white/70 px-3 py-2 text-xs font-medium transition hover:border-foreground/30"
                              >
                                <span>{item.label}</span>
                                <RadioGroupItem value={item.value} />
                              </label>
                            ))}
                          </RadioGroup>
                        </div>

                        {validationType === 'textMatching' && (
                          <div className="space-y-3">
                            <div className="grid gap-3 sm:grid-cols-2">
                              <div className="space-y-2">
                                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">匹配模式</p>
                                <RadioGroup
                                  value={validationConfig.textMatching.matchMode}
                                  onValueChange={(value) =>
                                    setValidationConfig((prev) => ({
                                      ...prev,
                                      textMatching: { ...prev.textMatching, matchMode: value },
                                    }))
                                  }
                                  className="grid grid-cols-2 gap-2"
                                >
                                  {['all', 'any'].map((item) => (
                                    <label
                                      key={item}
                                      className="flex items-center justify-between rounded-md border border-border/60 bg-white/70 px-3 py-2 text-xs font-medium"
                                    >
                                      <span>{item === 'all' ? '全部匹配' : '任意匹配'}</span>
                                      <RadioGroupItem value={item} />
                                    </label>
                                  ))}
                                </RadioGroup>
                              </div>
                              <div className="space-y-2">
                                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">区分大小写</p>
                                <label className="flex items-center gap-2 rounded-md border border-border/60 bg-white/70 px-3 py-3 text-sm">
                                  <Checkbox
                                    checked={validationConfig.textMatching.caseSensitive}
                                    onCheckedChange={(value) =>
                                      setValidationConfig((prev) => ({
                                        ...prev,
                                        textMatching: { ...prev.textMatching, caseSensitive: Boolean(value) },
                                      }))
                                    }
                                  />
                                  <span>大小写敏感</span>
                                </label>
                              </div>
                            </div>

                            <div className="space-y-2">
                              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">匹配文本</p>
                              <Textarea
                                value={textMatchingInput}
                                onChange={(event) => updateTextMatching(event.target.value)}
                                placeholder="每行一个匹配文本，例如 success"
                                className="min-h-[120px] bg-white/80"
                              />
                            </div>
                          </div>
                        )}

                        {validationType === 'lengthRange' && (
                          <div className="grid gap-3 sm:grid-cols-2">
                            <div className="space-y-2">
                              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">最小长度</p>
                              <Input
                                type="number"
                                min={0}
                                value={validationConfig.lengthRange.minLength}
                                onChange={(event) =>
                                  setValidationConfig((prev) => ({
                                    ...prev,
                                    lengthRange: {
                                      ...prev.lengthRange,
                                      minLength: Number(event.target.value || 0),
                                    },
                                  }))
                                }
                              />
                            </div>
                            <div className="space-y-2">
                              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">最大长度</p>
                              <Input
                                type="number"
                                min={-1}
                                value={validationConfig.lengthRange.maxLength}
                                onChange={(event) =>
                                  setValidationConfig((prev) => ({
                                    ...prev,
                                    lengthRange: {
                                      ...prev.lengthRange,
                                      maxLength: Number(event.target.value || 0),
                                    },
                                  }))
                                }
                              />
                            </div>
                          </div>
                        )}

                        <Separator />
                        <div className="grid gap-3 sm:grid-cols-2">
                          <div className="space-y-2">
                            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">超时时间(秒)</p>
                            <Input
                              type="number"
                              min={1}
                              max={300}
                              value={validationConfig.timeout}
                              onChange={(event) =>
                                setValidationConfig((prev) => ({
                                  ...prev,
                                  timeout: Number(event.target.value || 0),
                                }))
                              }
                            />
                          </div>
                          <div className="space-y-2">
                            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">重试次数</p>
                            <Input
                              type="number"
                              min={0}
                              max={10}
                              value={validationConfig.maxRetries}
                              onChange={(event) =>
                                setValidationConfig((prev) => ({
                                  ...prev,
                                  maxRetries: Number(event.target.value || 0),
                                }))
                              }
                            />
                          </div>
                        </div>
                      </div>
                    </>
                  )}
                </CardContent>
              </ScrollArea>

              {parsedRequest && (
                <CardFooter className="sticky-actions bg-white/90">
                  <div className="flex w-full flex-wrap items-center justify-between gap-3">
                    <Button
                      variant="secondary"
                      className="gap-2"
                      onClick={() => testRequestOnly()}
                      disabled={isSingleTesting}
                    >
                      {isSingleTesting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
                      {isSingleTesting ? '测试中...' : '测试请求'}
                    </Button>
                    <Button
                      variant="default"
                      className="gap-2"
                      onClick={testFieldNecessity}
                      disabled={!singleTestResult || isTestRunning}
                    >
                      {isTestRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <ScanSearch className="h-4 w-4" />}
                      字段分析
                    </Button>
                    <Button variant="outline" className="gap-2" onClick={clearResults} disabled={!parsedRequest}>
                      <RefreshCcw className="h-4 w-4" />
                      清空结果
                    </Button>
                  </div>
                </CardFooter>
              )}
            </Card>

            <ScrollArea className="h-[calc(100vh-240px)] rounded-xl">
              <div className="flex flex-col gap-6 pr-2">
                <Card className="glass-panel fade-in-up">
                  <CardHeader className="flex flex-row items-center justify-between gap-4">
                    <div>
                      <CardTitle className="section-title">Python 代码</CardTitle>
                      <CardDescription>生成可执行的请求脚本，支持快速复用。</CardDescription>
                    </div>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="outline"
                          size="sm"
                          className="gap-2"
                          onClick={() => copyCode(pythonCode)}
                          disabled={!pythonCode}
                        >
                          <Copy className="h-4 w-4" />
                          复制代码
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>复制生成的 Python 代码</TooltipContent>
                    </Tooltip>
                  </CardHeader>
                  <CardContent>
                    {pythonCode ? (
                      <pre className="code-block whitespace-pre-wrap">{pythonCode}</pre>
                    ) : (
                      <div className="rounded-xl border border-dashed border-border bg-white/70 px-6 py-10 text-center text-sm text-muted-foreground">
                        请先解析请求以生成 Python 代码
                      </div>
                    )}
                  </CardContent>
                </Card>

                <Card className="glass-panel fade-in-up" style={{ animationDelay: '80ms' }}>
                  <CardHeader>
                    <CardTitle className="section-title">请求测试结果</CardTitle>
                    <CardDescription>单次请求测试结果与响应细节。</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {singleTestResult ? (
                      <>
                        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                          <div className="metric-card">
                            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">状态码</p>
                            <Badge variant={badgeVariantByStatus(singleTestResult.statusCode)} className="mt-2">
                              {singleTestResult.statusCode}
                            </Badge>
                          </div>
                          <div className="metric-card">
                            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">响应大小</p>
                            <p className="mt-2 text-sm font-medium">{formatBytes(singleTestResult.contentLength || 0)}</p>
                          </div>
                          <div className="metric-card">
                            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">响应长度</p>
                            <p className="mt-2 text-sm font-medium">{singleTestResult.characterCount || 0} 字符</p>
                          </div>
                          <div className="metric-card">
                            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">耗时</p>
                            <p className="mt-2 text-sm font-medium">{formatDuration(singleTestResult.duration)}</p>
                          </div>
                        </div>

                        <div className="rounded-xl border border-border/60 bg-white/70 p-4 text-sm">
                          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">最终 URL</p>
                          <p className="mt-2 break-all font-medium text-foreground">{singleTestResult.url}</p>
                        </div>

                        {singleTestResult.detectedEncoding && (
                          <Badge variant="info" className="status-pill">
                            自动检测编码：{singleTestResult.detectedEncoding}
                          </Badge>
                        )}

                        <Tabs defaultValue="body">
                          <TabsList>
                            <TabsTrigger value="body">响应体</TabsTrigger>
                            <TabsTrigger value="headers">响应头</TabsTrigger>
                            {singleTestResult.cookies?.length > 0 && <TabsTrigger value="cookies">Cookies</TabsTrigger>}
                          </TabsList>
                          <TabsContent value="body" className="pt-2">
                            <div className="rounded-xl border border-border/60 bg-white/70 p-4">
                              <pre className="whitespace-pre-wrap text-xs leading-relaxed text-slate-700">{singleTestResult.body}</pre>
                            </div>
                          </TabsContent>
                          <TabsContent value="headers" className="pt-2">
                            <div className="rounded-xl border border-border/60 bg-white/70 p-3">
                              <Table>
                                <TableHeader>
                                  <TableRow>
                                    <TableHead className="w-[180px]">名称</TableHead>
                                    <TableHead>值</TableHead>
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  {formatHeaders(singleTestResult.headers).map((header) => (
                                    <TableRow key={header.name}>
                                      <TableCell className="font-medium">{header.name}</TableCell>
                                      <TableCell className="break-all text-muted-foreground">{header.value}</TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </div>
                          </TabsContent>
                          {singleTestResult.cookies?.length > 0 && (
                            <TabsContent value="cookies" className="pt-2">
                              <div className="rounded-xl border border-border/60 bg-white/70 p-3">
                                <Table>
                                  <TableHeader>
                                    <TableRow>
                                      <TableHead className="w-[140px]">名称</TableHead>
                                      <TableHead>值</TableHead>
                                      <TableHead className="w-[160px]">域名</TableHead>
                                      <TableHead className="w-[120px]">路径</TableHead>
                                    </TableRow>
                                  </TableHeader>
                                  <TableBody>
                                    {singleTestResult.cookies.map((cookie) => (
                                      <TableRow key={`${cookie.name}-${cookie.domain}`}> 
                                        <TableCell className="font-medium">{cookie.name}</TableCell>
                                        <TableCell className="break-all text-muted-foreground">{cookie.value}</TableCell>
                                        <TableCell className="text-muted-foreground">{cookie.domain}</TableCell>
                                        <TableCell className="text-muted-foreground">{cookie.path}</TableCell>
                                      </TableRow>
                                    ))}
                                  </TableBody>
                                </Table>
                              </div>
                            </TabsContent>
                          )}
                        </Tabs>
                      </>
                    ) : (
                      <div className="rounded-xl border border-dashed border-border bg-white/70 px-6 py-10 text-center text-sm text-muted-foreground">
                        请先解析请求进行测试
                      </div>
                    )}
                  </CardContent>
                </Card>

                <Card className="glass-panel fade-in-up" style={{ animationDelay: '160ms' }}>
                  <CardHeader>
                    <CardTitle className="section-title">测试摘要</CardTitle>
                    <CardDescription>字段必要性与整体通过率。</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {testResult ? (
                      <>
                        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                          <div className="metric-card">
                            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">原始请求</p>
                            <Badge variant={testResult.originalPassed ? 'success' : 'destructive'} className="mt-2">
                              {testResult.originalPassed ? '通过' : '失败'}
                            </Badge>
                          </div>
                          <div className="metric-card">
                            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">必需Headers</p>
                            <p className="mt-2 text-sm font-medium">
                              {testResultSummary?.requiredHeaders || 0} / {testResultSummary?.totalHeaders || 0}
                            </p>
                          </div>
                          <div className="metric-card">
                            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">必需Cookies</p>
                            <p className="mt-2 text-sm font-medium">
                              {testResultSummary?.requiredCookies || 0} / {testResultSummary?.totalCookies || 0}
                            </p>
                          </div>
                          <div className="metric-card">
                            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">测试耗时</p>
                            <p className="mt-2 text-sm font-medium">{formatDuration(testResult.testDuration)}</p>
                          </div>
                        </div>

                        {headerTestResults.length > 0 && (
                          <div className="rounded-xl border border-border/60 bg-white/70 p-3">
                            <div className="mb-3 flex items-center justify-between">
                              <p className="text-sm font-semibold">Headers 测试结果</p>
                              <Badge variant="outline">{headerTestResults.length} 条</Badge>
                            </div>
                            <Table>
                              <TableHeader>
                                <TableRow>
                                  <TableHead className="w-[200px]">Header名称</TableHead>
                                  <TableHead className="w-[120px]">是否必需</TableHead>
                                  <TableHead className="w-[100px]">状态码</TableHead>
                                  <TableHead>错误信息</TableHead>
                                </TableRow>
                              </TableHeader>
                              <TableBody>
                                {headerTestResults.map((row) => (
                                  <TableRow key={row.fieldName}>
                                    <TableCell className="font-medium">{row.fieldName}</TableCell>
                                    <TableCell>
                                      <Badge variant={row.isRequired ? 'destructive' : 'success'}>
                                        {row.isRequired ? '必需' : '可选'}
                                      </Badge>
                                    </TableCell>
                                    <TableCell>{row.statusCode}</TableCell>
                                    <TableCell className="text-muted-foreground">{row.errorMsg}</TableCell>
                                  </TableRow>
                                ))}
                              </TableBody>
                            </Table>
                          </div>
                        )}

                        {cookieTestResults.length > 0 && (
                          <div className="rounded-xl border border-border/60 bg-white/70 p-3">
                            <div className="mb-3 flex items-center justify-between">
                              <p className="text-sm font-semibold">Cookies 测试结果</p>
                              <Badge variant="outline">{cookieTestResults.length} 条</Badge>
                            </div>
                            <Table>
                              <TableHeader>
                                <TableRow>
                                  <TableHead className="w-[200px]">Cookie名称</TableHead>
                                  <TableHead className="w-[120px]">是否必需</TableHead>
                                  <TableHead className="w-[100px]">状态码</TableHead>
                                  <TableHead>错误信息</TableHead>
                                </TableRow>
                              </TableHeader>
                              <TableBody>
                                {cookieTestResults.map((row) => (
                                  <TableRow key={row.fieldName}>
                                    <TableCell className="font-medium">{row.fieldName}</TableCell>
                                    <TableCell>
                                      <Badge variant={row.isRequired ? 'destructive' : 'success'}>
                                        {row.isRequired ? '必需' : '可选'}
                                      </Badge>
                                    </TableCell>
                                    <TableCell>{row.statusCode}</TableCell>
                                    <TableCell className="text-muted-foreground">{row.errorMsg}</TableCell>
                                  </TableRow>
                                ))}
                              </TableBody>
                            </Table>
                          </div>
                        )}
                      </>
                    ) : (
                      <div className="rounded-xl border border-dashed border-border bg-white/70 px-6 py-10 text-center text-sm text-muted-foreground">
                        请先进行字段分析测试
                      </div>
                    )}
                  </CardContent>
                </Card>

                <Card className="glass-panel fade-in-up" style={{ animationDelay: '240ms' }}>
                  <CardHeader className="flex flex-row items-center justify-between gap-4">
                    <div>
                      <CardTitle className="section-title">简化代码</CardTitle>
                      <CardDescription>字段必要性分析后生成的精简版本。</CardDescription>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      className="gap-2"
                      onClick={() => copyCode(testResult?.simplifiedCode)}
                      disabled={!testResult?.simplifiedCode}
                    >
                      <Copy className="h-4 w-4" />
                      复制代码
                    </Button>
                  </CardHeader>
                  <CardContent>
                    {testResult?.simplifiedCode ? (
                      <pre className="code-block whitespace-pre-wrap">{testResult.simplifiedCode}</pre>
                    ) : (
                      <div className="rounded-xl border border-dashed border-border bg-white/70 px-6 py-10 text-center text-sm text-muted-foreground">
                        完成字段分析后将显示简化代码
                      </div>
                    )}
                  </CardContent>
                </Card>
              </div>
            </ScrollArea>
          </div>
        </div>
      </div>
    </TooltipProvider>
  );
}
