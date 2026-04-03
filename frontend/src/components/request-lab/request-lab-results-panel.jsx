import React from 'react';

import { Badge } from '../ui/badge.jsx';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table.jsx';
import { CodeBlockCard } from '../shared/code-block-card.jsx';
import { ContainerCard } from '../shared/container-card.jsx';
import { EmptyState } from '../shared/empty-state.jsx';
import { MetricCard } from '../shared/metric-card.jsx';
import { TabNavigator } from '../shared/tab-navigator.jsx';

function ResultDetailsTabs({ singleTestResult, formatHeaders }) {
  const tabs = [
    {
      value: 'body',
      label: '响应体',
      content: (
        <div className="rounded-xl border border-border/60 bg-white/70 p-4">
          <pre className="whitespace-pre-wrap break-words [overflow-wrap:anywhere] text-xs leading-relaxed text-slate-700">
            {singleTestResult.body}
          </pre>
        </div>
      ),
    },
    {
      value: 'headers',
      label: '响应头',
      content: (
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
                  <TableCell className="break-all font-medium [overflow-wrap:anywhere]">{header.name}</TableCell>
                  <TableCell className="break-all text-muted-foreground">{header.value}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      ),
    },
  ];

  if (singleTestResult.cookies?.length > 0) {
    tabs.push({
      value: 'cookies',
      label: 'Cookies',
      content: (
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
                  <TableCell className="break-all font-medium [overflow-wrap:anywhere]">{cookie.name}</TableCell>
                  <TableCell className="break-all text-muted-foreground">{cookie.value}</TableCell>
                  <TableCell className="break-all text-muted-foreground [overflow-wrap:anywhere]">{cookie.domain}</TableCell>
                  <TableCell className="break-all text-muted-foreground [overflow-wrap:anywhere]">{cookie.path}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      ),
    });
  }

  return (
    <TabNavigator
      defaultValue="body"
      tabs={tabs}
      className="min-h-0 min-w-0"
      listClassName="max-w-full"
      contentClassName="min-h-0 flex-1 pt-2"
    />
  );
}

function RequestTestGroup({
  pythonCode,
  singleTestResult,
  onCopyCode,
  badgeVariantByStatus,
  formatBytes,
  formatDuration,
  formatHeaders,
}) {
  return (
    <div className="space-y-6">
      <CodeBlockCard
        title="Python 代码"
        description="生成可执行的请求脚本，支持快速复用。"
        code={pythonCode}
        emptyMessage="请先解析请求以生成 Python 代码"
        onCopy={onCopyCode}
        className="fade-in-up"
      />

      <ContainerCard
        title="请求测试结果"
        description="单次请求测试结果与响应细节。"
        className="fade-in-up"
        style={{ animationDelay: '80ms' }}
        contentClassName="space-y-4"
      >
        {singleTestResult ? (
          <>
            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <MetricCard label="状态码">
                <Badge variant={badgeVariantByStatus(singleTestResult.statusCode)}>
                  {singleTestResult.statusCode}
                </Badge>
              </MetricCard>
              <MetricCard label="响应大小">
                <p className="text-sm font-medium">{formatBytes(singleTestResult.contentLength || 0)}</p>
              </MetricCard>
              <MetricCard label="响应长度">
                <p className="text-sm font-medium">{singleTestResult.characterCount || 0} 字符</p>
              </MetricCard>
              <MetricCard label="耗时">
                <p className="text-sm font-medium">{formatDuration(singleTestResult.duration)}</p>
              </MetricCard>
            </div>

            <div className="rounded-xl border border-border/60 bg-white/70 p-4 text-sm">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">最终 URL</p>
              <p className="mt-2 break-all font-medium text-foreground">{singleTestResult.url}</p>
            </div>

            {singleTestResult.detectedEncoding ? (
              <Badge variant="info" className="status-pill">
                自动检测编码：{singleTestResult.detectedEncoding}
              </Badge>
            ) : null}

            <ResultDetailsTabs singleTestResult={singleTestResult} formatHeaders={formatHeaders} />
          </>
        ) : (
          <EmptyState message="请先解析请求进行测试" />
        )}
      </ContainerCard>
    </div>
  );
}

function AnalysisGroup({
  testResult,
  testResultSummary,
  headerTestResults,
  cookieTestResults,
  onCopyCode,
  formatDuration,
}) {
  return (
    <div className="space-y-6">
      <ContainerCard
        title="测试摘要"
        description="字段必要性与整体通过率。"
        className="fade-in-up"
        style={{ animationDelay: '160ms' }}
        contentClassName="space-y-4"
      >
        {testResult ? (
          <>
            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <MetricCard label="原始请求">
                <Badge variant={testResult.originalPassed ? 'success' : 'destructive'}>
                  {testResult.originalPassed ? '通过' : '失败'}
                </Badge>
              </MetricCard>
              <MetricCard label="必需Headers">
                <p className="text-sm font-medium">
                  {testResultSummary?.requiredHeaders || 0} / {testResultSummary?.totalHeaders || 0}
                </p>
              </MetricCard>
              <MetricCard label="必需Cookies">
                <p className="text-sm font-medium">
                  {testResultSummary?.requiredCookies || 0} / {testResultSummary?.totalCookies || 0}
                </p>
              </MetricCard>
              <MetricCard label="测试耗时">
                <p className="text-sm font-medium">{formatDuration(testResult.testDuration)}</p>
              </MetricCard>
            </div>

            {headerTestResults.length > 0 ? (
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
                        <TableCell className="break-all font-medium [overflow-wrap:anywhere]">{row.fieldName}</TableCell>
                        <TableCell>
                          <Badge variant={row.isRequired ? 'destructive' : 'success'}>
                            {row.isRequired ? '必需' : '可选'}
                          </Badge>
                        </TableCell>
                        <TableCell>{row.statusCode}</TableCell>
                        <TableCell className="break-all text-muted-foreground [overflow-wrap:anywhere]">{row.errorMsg}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : null}

            {cookieTestResults.length > 0 ? (
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
                        <TableCell className="break-all font-medium [overflow-wrap:anywhere]">{row.fieldName}</TableCell>
                        <TableCell>
                          <Badge variant={row.isRequired ? 'destructive' : 'success'}>
                            {row.isRequired ? '必需' : '可选'}
                          </Badge>
                        </TableCell>
                        <TableCell>{row.statusCode}</TableCell>
                        <TableCell className="break-all text-muted-foreground [overflow-wrap:anywhere]">{row.errorMsg}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : null}
          </>
        ) : (
          <EmptyState message="请先进行字段分析测试" />
        )}
      </ContainerCard>

      <CodeBlockCard
        title="简化代码"
        description="字段必要性分析后生成的精简版本。"
        code={testResult?.simplifiedCode}
        emptyMessage="完成字段分析后将显示简化代码"
        onCopy={onCopyCode}
        className="fade-in-up"
        style={{ animationDelay: '240ms' }}
      />
    </div>
  );
}

export function RequestLabResultsPanel({
  activeRightPanelTab,
  onRightPanelTabChange,
  pythonCode,
  singleTestResult,
  testResult,
  testResultSummary,
  headerTestResults,
  cookieTestResults,
  onCopyCode,
  badgeVariantByStatus,
  formatBytes,
  formatDuration,
  formatHeaders,
}) {
  const tabs = [
    {
      value: 'request-test-group',
      label: '请求代码与测试',
      content: (
        <RequestTestGroup
          pythonCode={pythonCode}
          singleTestResult={singleTestResult}
          onCopyCode={onCopyCode}
          badgeVariantByStatus={badgeVariantByStatus}
          formatBytes={formatBytes}
          formatDuration={formatDuration}
          formatHeaders={formatHeaders}
        />
      ),
      contentClassName: 'space-y-6',
    },
    {
      value: 'analysis-group',
      label: '测试摘要与简化代码',
      content: (
        <AnalysisGroup
          testResult={testResult}
          testResultSummary={testResultSummary}
          headerTestResults={headerTestResults}
          cookieTestResults={cookieTestResults}
          onCopyCode={onCopyCode}
          formatDuration={formatDuration}
        />
      ),
      contentClassName: 'space-y-6',
    },
  ];

  return (
    <TabNavigator
      value={activeRightPanelTab}
      onValueChange={onRightPanelTabChange}
      tabs={tabs}
      className="flex h-full min-h-0 min-w-0 flex-col gap-4 pr-1"
      listClassName="max-w-full self-start"
      contentClassName="mt-0 min-h-0 flex-1 overflow-y-auto pr-1"
    />
  );
}
