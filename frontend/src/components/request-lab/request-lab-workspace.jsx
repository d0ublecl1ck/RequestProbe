import React, { useEffect, useMemo, useState } from 'react';

import { toast } from 'sonner';

import {
  DetectInputType,
  GeneratePythonCode,
  GetDefaultValidationConfig,
  GetSupportedEncodings,
  ParseRequest,
  ParseRequestWithType,
  TestFieldNecessity,
  TestRequestOnly,
} from '../../../wailsjs/go/main/App.js';
import { Badge } from '../ui/badge.jsx';
import { RequestLabInputPanel } from './request-lab-input-panel.jsx';
import { RequestLabResultsPanel } from './request-lab-results-panel.jsx';

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

export function RequestLabWorkspace() {
  const [activeRightPanelTab, setActiveRightPanelTab] = useState('request-test-group');
  const [inputText, setInputText] = useState('');
  const [inputType, setInputType] = useState('auto');
  const [detectedType, setDetectedType] = useState('');
  const [isParsing, setIsParsing] = useState(false);
  const [parsedRequest, setParsedRequest] = useState(null);
  const [pythonCode, setPythonCode] = useState('');
  const [validationType, setValidationType] = useState('textMatching');
  const [validationConfig, setValidationConfig] = useState(defaultValidationConfig);
  const [textMatchingInput, setTextMatchingInput] = useState('');
  const [testResult, setTestResult] = useState(null);
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
            supportedEncodings:
              encodings ||
              config.encodingConfig?.supportedEncodings ||
              defaultValidationConfig.encodingConfig.supportedEncodings,
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

      const code = await GeneratePythonCode(request);
      setPythonCode(code);

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
      setActiveRightPanelTab('analysis-group');
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
    setTextMatchingInput('');
    setValidationConfig((prev) => ({
      ...prev,
      textMatching: {
        ...prev.textMatching,
        texts: [],
      },
    }));
    setTestResult(null);
    setSingleTestResult(null);
    setActiveRightPanelTab('request-test-group');
    toast.success('所有内容已清空');
  };

  const copyCode = async (code) => {
    if (!code) return;
    try {
      await navigator.clipboard.writeText(code);
      toast.success('代码已复制到剪贴板');
    } catch {
      toast.error('复制失败');
    }
  };

  const heroMetrics = [
    {
      label: '输入模式',
      value: inputType === 'auto' ? '自动检测' : inputType === 'raw' ? 'Raw HTTP' : 'Curl 命令',
      copy: detectedType ? `已识别为 ${detectedType === 'curl' ? 'Curl' : detectedType === 'raw' ? 'Raw HTTP' : '未知格式'}` : '粘贴请求后自动推断格式',
    },
    {
      label: '当前阶段',
      value: parsedRequest ? '已完成解析' : '等待输入',
      copy: parsedRequest ? '可以继续进行单次测试或字段分析' : '先输入请求文本，再触发解析',
    },
    {
      label: '结果焦点',
      value: activeRightPanelTab === 'analysis-group' ? '测试摘要' : '代码与测试',
      copy: testResult ? '最近一次字段分析结果已生成' : '右侧负责承载测试与简化代码',
    },
  ];

  return (
    <div className="workspace-page">
      <section className="workspace-hero">
        <p className="workspace-kicker">字段探针</p>
        <h1 className="workspace-hero-title">请求分析</h1>
        <p className="workspace-hero-copy">
          粘贴原始请求后，依次完成解析、单次测试和字段必要性分析，并输出可复用的 Python 代码。
        </p>
        <div className="workspace-hero-grid">
          {heroMetrics.map((item) => (
            <div key={item.label} className="workspace-hero-metric">
              <p className="workspace-hero-metric-label">{item.label}</p>
              <p className="workspace-hero-metric-value">{item.value}</p>
              <p className="workspace-hero-metric-copy">{item.copy}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="workspace-section">
        <div className="workspace-section-header">
          <div>
            <h2 className="workspace-section-title">分析工作台</h2>
            <p className="workspace-section-copy">
              左侧配置输入和验证条件，右侧查看代码、测试结果和简化脚本。
            </p>
          </div>
          <Badge variant={parsedRequest ? 'success' : 'outline'}>
            {parsedRequest ? '解析完成' : '待输入'}
          </Badge>
        </div>

        <div className="editorial-grid editorial-grid-2 p-4 xl:p-5">
          <RequestLabInputPanel
            inputType={inputType}
            onInputTypeChange={setInputType}
            detectedType={detectedType}
            inputText={inputText}
            onInputTextChange={setInputText}
            onParseRequest={parseRequest}
            isParsing={isParsing}
            parsedRequest={parsedRequest}
            validationType={validationType}
            onValidationTypeChange={updateValidationType}
            validationConfig={validationConfig}
            onValidationConfigChange={setValidationConfig}
            textMatchingInput={textMatchingInput}
            onTextMatchingChange={updateTextMatching}
            onSingleTest={() => testRequestOnly()}
            isSingleTesting={isSingleTesting}
            onFieldAnalysis={testFieldNecessity}
            isTestRunning={isTestRunning}
            onClearAll={clearResults}
          />

          <div className="flex min-h-[420px] min-w-0 flex-col xl:min-h-0">
            <RequestLabResultsPanel
              activeRightPanelTab={activeRightPanelTab}
              onRightPanelTabChange={setActiveRightPanelTab}
              pythonCode={pythonCode}
              singleTestResult={singleTestResult}
              testResult={testResult}
              testResultSummary={testResultSummary}
              headerTestResults={headerTestResults}
              cookieTestResults={cookieTestResults}
              onCopyCode={copyCode}
              badgeVariantByStatus={badgeVariantByStatus}
              formatBytes={formatBytes}
              formatDuration={formatDuration}
              formatHeaders={formatHeaders}
            />
          </div>
        </div>
      </section>
    </div>
  );
}
