import React from 'react';
import { Loader2, Play, RefreshCcw, ScanSearch, Trash2, Wand2 } from 'lucide-react';

import { Button } from '../ui/button.jsx';
import { Input } from '../ui/input.jsx';
import { ScrollArea } from '../ui/scroll-area.jsx';
import { Separator } from '../ui/separator.jsx';
import { Textarea } from '../ui/textarea.jsx';
import { Badge } from '../ui/badge.jsx';
import { RadioGroup, RadioGroupItem } from '../ui/radio-group.jsx';
import { CardFooter } from '../ui/card.jsx';
import { CheckboxCard } from '../shared/checkbox-card.jsx';
import { ContainerCard } from '../shared/container-card.jsx';
import { FormSection } from '../shared/form-section.jsx';
import { RadioCardGroup } from '../shared/radio-card-group.jsx';

const INPUT_TYPE_OPTIONS = [
  { value: 'auto', label: '自动检测' },
  { value: 'raw', label: 'Raw HTTP' },
  { value: 'curl', label: 'Curl 命令' },
];

const VALIDATION_TYPE_OPTIONS = [
  { value: 'textMatching', label: '文本匹配' },
  { value: 'lengthRange', label: '响应长度' },
];

const MATCH_MODE_OPTIONS = [
  { value: 'all', label: '全部匹配' },
  { value: 'any', label: '任意匹配' },
];

export function RequestLabInputPanel({
  inputType,
  onInputTypeChange,
  detectedType,
  inputText,
  onInputTextChange,
  onParseRequest,
  isParsing,
  parsedRequest,
  validationType,
  onValidationTypeChange,
  validationConfig,
  onValidationConfigChange,
  textMatchingInput,
  onTextMatchingChange,
  onSingleTest,
  isSingleTesting,
  onFieldAnalysis,
  isTestRunning,
  onClearAll,
}) {
  return (
    <ContainerCard
      title="请求输入"
      icon={ScanSearch}
      className="flex min-h-[420px] min-w-0 flex-col xl:h-full xl:min-h-0"
      contentClassName="flex min-h-0 flex-1 flex-col p-0"
    >
      <ScrollArea className="min-h-0 flex-1">
        <div className="flex flex-col gap-6 p-6">
          <FormSection title="输入类型">
            <RadioCardGroup
              value={inputType}
              onValueChange={onInputTypeChange}
              options={INPUT_TYPE_OPTIONS}
              className="grid grid-cols-3 gap-2"
            />
          </FormSection>

          <FormSection title="请求内容">
            <div className="flex items-center justify-between">
              <div />
              {detectedType ? (
                <Badge variant="outline">
                  检测到：{detectedType === 'curl' ? 'Curl' : detectedType === 'raw' ? 'Raw HTTP' : '未知格式'}
                </Badge>
              ) : null}
            </div>
            <Textarea
              value={inputText}
              onChange={(event) => onInputTextChange(event.target.value)}
              placeholder="请输入HTTP请求或Curl命令..."
              className="min-h-[220px] resize-none bg-white/80"
            />
          </FormSection>

          <div className="grid gap-3 sm:grid-cols-2">
            <Button variant="default" className="gap-2" onClick={onParseRequest} disabled={isParsing}>
              {isParsing ? <Loader2 className="h-4 w-4 animate-spin" /> : <Wand2 className="h-4 w-4" />}
              {isParsing ? '解析中...' : '解析请求'}
            </Button>
            <Button
              variant="outline"
              className="gap-2"
              onClick={() => onInputTextChange('')}
              disabled={!inputText}
            >
              <Trash2 className="h-4 w-4" />
              清空输入
            </Button>
          </div>

          {parsedRequest ? (
            <>
              <Separator />
              <div className="space-y-4">
                <div>
                  <p className="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">验证配置</p>
                </div>

                <FormSection title="验证方式">
                  <RadioCardGroup
                    value={validationType}
                    onValueChange={onValidationTypeChange}
                    options={VALIDATION_TYPE_OPTIONS}
                    className="grid grid-cols-2 gap-2"
                  />
                </FormSection>

                {validationType === 'textMatching' ? (
                  <FormSection title="文本匹配">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <div className="space-y-2">
                        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">匹配模式</p>
                        <RadioGroup value={validationConfig.textMatching.matchMode} onValueChange={(value) => onValidationConfigChange((prev) => ({
                          ...prev,
                          textMatching: { ...prev.textMatching, matchMode: value },
                        }))} className="grid grid-cols-2 gap-2">
                          {MATCH_MODE_OPTIONS.map((option) => (
                            <label
                              key={option.value}
                              className="flex items-center justify-between rounded-md border border-border/60 bg-white/70 px-3 py-2 text-xs font-medium"
                            >
                              <span>{option.label}</span>
                              <RadioGroupItem value={option.value} />
                            </label>
                          ))}
                        </RadioGroup>
                      </div>
                      <div className="space-y-2">
                        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">区分大小写</p>
                        <CheckboxCard
                          checked={validationConfig.textMatching.caseSensitive}
                          onCheckedChange={(checked) => onValidationConfigChange((prev) => ({
                            ...prev,
                            textMatching: { ...prev.textMatching, caseSensitive: Boolean(checked) },
                          }))}
                          label="大小写敏感"
                        />
                      </div>
                    </div>
                    <div className="space-y-2">
                      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">匹配文本</p>
                      <Textarea
                        value={textMatchingInput}
                        onChange={(event) => onTextMatchingChange(event.target.value)}
                        placeholder="每行一个匹配文本，例如 success"
                        className="min-h-[120px] bg-white/80"
                      />
                    </div>
                  </FormSection>
                ) : null}

                {validationType === 'lengthRange' ? (
                  <FormSection title="响应长度">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <div className="space-y-2">
                        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">最小长度</p>
                        <Input
                          type="number"
                          min={0}
                          value={validationConfig.lengthRange.minLength}
                          onChange={(event) => onValidationConfigChange((prev) => ({
                            ...prev,
                            lengthRange: {
                              ...prev.lengthRange,
                              minLength: Number(event.target.value || 0),
                            },
                          }))}
                        />
                      </div>
                      <div className="space-y-2">
                        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">最大长度</p>
                        <Input
                          type="number"
                          min={-1}
                          value={validationConfig.lengthRange.maxLength}
                          onChange={(event) => onValidationConfigChange((prev) => ({
                            ...prev,
                            lengthRange: {
                              ...prev.lengthRange,
                              maxLength: Number(event.target.value || 0),
                            },
                          }))}
                        />
                      </div>
                    </div>
                  </FormSection>
                ) : null}

                <Separator />

                <FormSection title="请求参数">
                  <div className="grid gap-3 sm:grid-cols-2">
                    <div className="space-y-2">
                      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">超时时间(秒)</p>
                      <Input
                        type="number"
                        min={1}
                        max={300}
                        value={validationConfig.timeout}
                        onChange={(event) => onValidationConfigChange((prev) => ({
                          ...prev,
                          timeout: Number(event.target.value || 0),
                        }))}
                      />
                    </div>
                    <div className="space-y-2">
                      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">重试次数</p>
                      <Input
                        type="number"
                        min={0}
                        max={10}
                        value={validationConfig.maxRetries}
                        onChange={(event) => onValidationConfigChange((prev) => ({
                          ...prev,
                          maxRetries: Number(event.target.value || 0),
                        }))}
                      />
                    </div>
                  </div>
                </FormSection>
              </div>
            </>
          ) : null}
        </div>
      </ScrollArea>

      {parsedRequest ? (
        <CardFooter className="sticky-actions bg-white/90">
          <div className="grid w-full gap-3 md:grid-cols-2 xl:grid-cols-3">
            <Button variant="secondary" className="gap-2" onClick={onSingleTest} disabled={isSingleTesting}>
              {isSingleTesting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
              {isSingleTesting ? '测试中...' : '测试请求'}
            </Button>
            <Button
              variant="default"
              className="gap-2"
              onClick={onFieldAnalysis}
              disabled={!parsedRequest || isTestRunning}
            >
              {isTestRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <ScanSearch className="h-4 w-4" />}
              字段分析
            </Button>
            <Button variant="outline" className="gap-2" onClick={onClearAll} disabled={!parsedRequest}>
              <RefreshCcw className="h-4 w-4" />
              清空结果
            </Button>
          </div>
        </CardFooter>
      ) : null}
    </ContainerCard>
  );
}
