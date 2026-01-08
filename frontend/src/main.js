import './style.css';
import './app.css';

// 导入Wails运行时
import {
  Greet,
  ParseRequest,
  ParseRequestWithType,
  DetectInputType,
  GeneratePythonCode,
  TestSingleRequest,
  TestFieldNecessity,
  ValidateExpression,

  GetDefaultValidationConfig,
  GetRequestSummary,
  GetTestStatistics,
  TestRequestOnly,
  GetSupportedEncodings,


  ParseSimpleRequest
} from '../wailsjs/go/main/App.js';

const { createApp, ref, reactive, computed, onMounted, nextTick } = Vue;
const { ElMessage, ElMessageBox, ElLoading } = ElementPlus;

const app = createApp({
  template: `
    <div class="app-container">


      <!-- 主要内容区域 -->
      <div class="main-content">
        <!-- 左侧面板 -->
        <div class="left-panel" :style="{ width: state.leftPanelWidth }">
          <!-- 滚动内容区域 -->
          <div class="left-panel-scroll">
            <el-card class="input-card" shadow="never">
              <template #header>
                <div class="card-header">
                  <span>请求输入</span>
                  <div class="input-type-selector">
                    <el-radio-group v-model="state.inputType" size="small">
                      <el-radio-button label="auto">自动检测</el-radio-button>
                      <el-radio-button label="raw">Raw HTTP</el-radio-button>
                      <el-radio-button label="curl">Curl命令</el-radio-button>
                    </el-radio-group>
                  </div>
                </div>
              </template>

              <div class="input-section">
                <el-input
                  v-model="state.inputText"
                  type="textarea"
                  :rows="12"
                  placeholder="请输入HTTP请求或Curl命令..."
                  @input="detectInputType"
                  class="input-textarea"
                />

                <div class="input-info" v-if="state.detectedType">
                  <el-tag size="small" type="info">
                    检测到: {{ state.detectedType === 'curl' ? 'Curl命令' : state.detectedType === 'raw' ? 'Raw HTTP' : '未知格式' }}
                  </el-tag>
                </div>

                <div class="input-actions">
                  <el-button type="primary" @click="parseRequest" :disabled="state.isParsing">
                    <el-icon v-if="!state.isParsing"><DocumentCopy /></el-icon>
                    <el-icon v-else class="is-loading"><Loading /></el-icon>
                    {{ state.isParsing ? '解析中...' : '解析请求' }}
                  </el-button>
                  <el-button @click="state.inputText = ''" :disabled="!state.inputText">
                    <el-icon><Delete /></el-icon>
                    清空
                  </el-button>
                </div>
              </div>
            </el-card>

            <!-- 验证配置 -->
            <el-card class="config-card" shadow="never" v-if="state.parsedRequest">
              <template #header>
                <span>验证配置</span>
              </template>

              <div class="config-section">
                <!-- 验证方式选择 -->
                <div class="config-item">
                  <div class="config-header">
                    <label>验证方式:</label>
                  </div>
                  <el-radio-group v-model="state.validationType" size="small" @change="onValidationTypeChange">
                    <el-radio-button label="textMatching">文本匹配验证</el-radio-button>
                    <el-radio-button label="lengthRange">响应长度验证</el-radio-button>
                  </el-radio-group>
                </div>

                <!-- 文本匹配验证配置 -->
                <div v-if="state.validationType === 'textMatching'" class="config-sub-section">
                  <div class="config-row">
                    <div class="config-item">
                      <label>匹配模式:</label>
                      <el-radio-group v-model="state.validationConfig.textMatching.matchMode" size="small">
                        <el-radio-button label="all">全部匹配</el-radio-button>
                        <el-radio-button label="any">任意匹配</el-radio-button>
                      </el-radio-group>
                    </div>
                    <div class="config-item">
                      <el-checkbox v-model="state.validationConfig.textMatching.caseSensitive">
                        区分大小写
                      </el-checkbox>
                    </div>
                  </div>
                  <div class="config-item">
                    <label>匹配文本 (每行一个):</label>
                    <el-input
                      ref="textMatchingInput"
                      v-model="state.textMatchingInput"
                      type="textarea"
                      :autosize="{ minRows: 1, maxRows: 10 }"
                      placeholder="输入要匹配的文本，每行一个&#10;例如：success"
                      @input="updateTextMatching"
                    />
                  </div>
                </div>

                <!-- 响应长度验证配置 -->
                <div v-if="state.validationType === 'lengthRange'" class="config-sub-section">
                  <div class="config-row">
                    <div class="config-item">
                      <label>最小长度:</label>
                      <el-input-number
                        v-model="state.validationConfig.lengthRange.minLength"
                        :min="0"
                        size="small"
                      />
                    </div>
                    <div class="config-item">
                      <label>最大长度:</label>
                      <el-input-number
                        v-model="state.validationConfig.lengthRange.maxLength"
                        :min="-1"
                        size="small"
                        placeholder="-1表示无限制"
                      />
                    </div>
                  </div>
                </div>

                <!-- 高级配置 -->
                <div class="config-row">
                  <div class="config-item">
                    <label>超时时间(秒):</label>
                    <el-input-number
                      v-model="state.validationConfig.timeout"
                      :min="1"
                      :max="300"
                      size="small"
                    />
                  </div>
                  <div class="config-item">
                    <label>重试次数:</label>
                    <el-input-number
                      v-model="state.validationConfig.maxRetries"
                      :min="0"
                      :max="10"
                      size="small"
                    />
                  </div>
                </div>
              </div>
            </el-card>
          </div>

          <!-- 固定在底部的操作按钮 -->
          <div class="left-panel-actions" v-if="state.parsedRequest">
            <!-- 字段分析（只有测试请求成功后才显示） -->
            <el-button
              v-if="state.singleTestResult"
              type="success"
              @click="testFieldNecessity"
              :disabled="state.isTestRunning"
              size="large"
            >
              <el-icon v-if="!state.isTestRunning"><VideoPlay /></el-icon>
              <el-icon v-else class="is-loading"><Loading /></el-icon>
              {{ state.isTestRunning ? '分析中...' : '字段分析' }}
            </el-button>

            <!-- 清空按钮（只有完成分析后才显示） -->
            <el-button
              v-if="state.testResult"
              type="warning"
              @click="clearResults"
              size="large"
            >
              <el-icon><RefreshLeft /></el-icon>
              清空结果
            </el-button>
          </div>
        </div>

        <!-- 右侧面板 -->
        <div class="right-panel">
          <div class="content-scroll">
            <!-- Python代码区域 -->
            <div class="section-card">
              <div class="section-header">
                <h3>Python代码</h3>
                <el-button size="small" @click="copyCode(state.pythonCode)" :disabled="!state.pythonCode">
                  <el-icon><DocumentCopy /></el-icon>
                  复制代码
                </el-button>
              </div>
              <div class="code-content">
                <pre class="code-block" v-if="state.pythonCode">{{ state.pythonCode }}</pre>
                <div v-else style="text-align: center; color: #909399; padding: 40px; border: 2px dashed #ddd; border-radius: 8px; background: #f9f9f9;">
                  请先解析请求以生成Python代码
                </div>
              </div>
            </div>

            <!-- 请求测试区域 -->
            <div class="section-card">
              <div class="section-header">
                <h3>请求测试结果</h3>
              </div>
              <div v-if="state.singleTestResult">
                <!-- 测试摘要 -->
                <div class="test-summary-grid">
                  <div class="summary-item">
                    <span class="label">状态码:</span>
                    <el-tag :type="getStatusCodeType(state.singleTestResult.statusCode)">
                      {{ state.singleTestResult.statusCode }}
                    </el-tag>
                  </div>
                  <div class="summary-item">
                    <span class="label">响应大小:</span>
                    <span class="value">{{ formatBytes(state.singleTestResult.contentLength || 0) }}</span>
                  </div>
                  <div class="summary-item">
                    <span class="label">响应长度:</span>
                    <span class="value">{{ state.singleTestResult.characterCount || 0 }} 字符</span>
                  </div>
                  <div class="summary-item">
                    <span class="label">请求耗时:</span>
                    <span class="value">{{ formatDuration(state.singleTestResult.duration) }}</span>
                  </div>
                </div>

                <!-- 最终URL -->
                <div class="url-section">
                  <span class="label">最终URL:</span>
                  <div class="url-text">{{ state.singleTestResult.url }}</div>
                </div>

                <!-- 编码信息 -->
                <div class="encoding-info" v-if="state.singleTestResult.detectedEncoding">
                  <el-tag type="info" size="small">
                    自动检测编码: {{ state.singleTestResult.detectedEncoding }}
                  </el-tag>
                </div>

                <!-- 响应内容 -->
                <div class="response-content">
                  <el-tabs type="border-card">
                    <el-tab-pane label="响应体">
                      <div class="response-body">
                        <pre class="response-text">{{ state.singleTestResult.body }}</pre>
                      </div>
                    </el-tab-pane>
                    <el-tab-pane label="响应头">
                      <el-table :data="formatHeaders(state.singleTestResult.headers)" size="small">
                        <el-table-column prop="name" label="名称" width="200" />
                        <el-table-column prop="value" label="值" show-overflow-tooltip />
                      </el-table>
                    </el-tab-pane>
                    <el-tab-pane label="Cookies" v-if="state.singleTestResult.cookies?.length > 0">
                      <el-table :data="state.singleTestResult.cookies" size="small">
                        <el-table-column prop="name" label="名称" width="150" />
                        <el-table-column prop="value" label="值" show-overflow-tooltip />
                        <el-table-column prop="domain" label="域名" width="150" />
                        <el-table-column prop="path" label="路径" width="100" />
                      </el-table>
                    </el-tab-pane>
                  </el-tabs>
                </div>
              </div>
              <div v-else class="placeholder-content">
                请先解析请求进行测试
              </div>
            </div>

            <!-- 测试结果区域 -->
            <div class="section-card">
              <div class="section-header">
                <h3>测试摘要</h3>
              </div>
              <div class="summary-content" v-if="state.testResult">
                    <div class="summary-item">
                      <span class="label">原始请求:</span>
                      <el-tag :type="state.testResult.originalPassed ? 'success' : 'danger'">
                        {{ state.testResult.originalPassed ? '通过' : '失败' }}
                      </el-tag>
                    </div>
                    <div class="summary-item">
                      <span class="label">必需Headers:</span>
                      <span>{{ testResultSummary?.requiredHeaders || 0 }} / {{ testResultSummary?.totalHeaders || 0 }}</span>
                    </div>
                    <div class="summary-item">
                      <span class="label">必需Cookies:</span>
                      <span>{{ testResultSummary?.requiredCookies || 0 }} / {{ testResultSummary?.totalCookies || 0 }}</span>
                    </div>
                    <div class="summary-item">
                      <span class="label">测试耗时:</span>
                      <span>{{ formatDuration(state.testResult.testDuration) }}</span>
                    </div>
                </div>

                <!-- Headers测试结果表格 -->
                <el-card class="table-card" shadow="never" v-if="headerTestResults.length > 0">
                  <template #header>
                    <span>Headers测试结果</span>
                  </template>

                  <el-table :data="headerTestResults" stripe>
                    <el-table-column prop="fieldName" label="Header名称" width="200" />
                    <el-table-column prop="isRequired" label="是否必需" width="100">
                      <template #default="scope">
                        <el-tag :type="scope.row.isRequired ? 'danger' : 'success'" size="small">
                          {{ scope.row.isRequired ? '必需' : '可选' }}
                        </el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column prop="statusCode" label="状态码" width="100" />
                    <el-table-column prop="errorMsg" label="错误信息" show-overflow-tooltip />
                  </el-table>
                </el-card>

                <!-- Cookies测试结果表格 -->
                <el-card class="table-card" shadow="never" v-if="cookieTestResults.length > 0">
                  <template #header>
                    <span>Cookies测试结果</span>
                  </template>

                  <el-table :data="cookieTestResults" stripe>
                    <el-table-column prop="fieldName" label="Cookie名称" width="200" />
                    <el-table-column prop="isRequired" label="是否必需" width="100">
                      <template #default="scope">
                        <el-tag :type="scope.row.isRequired ? 'danger' : 'success'" size="small">
                          {{ scope.row.isRequired ? '必需' : '可选' }}
                        </el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column prop="statusCode" label="状态码" width="100" />
                    <el-table-column prop="errorMsg" label="错误信息" show-overflow-tooltip />
                  </el-table>
                </el-card>
              <div v-else class="placeholder-content">
                请先进行字段分析测试
              </div>
            </div>

            <!-- 简化代码区域 -->
            <div class="section-card">
              <div class="section-header">
                <h3>简化代码</h3>
                <el-button size="small" @click="copyCode(state.testResult?.simplifiedCode)" :disabled="!state.testResult?.simplifiedCode">
                  <el-icon><DocumentCopy /></el-icon>
                  复制代码
                </el-button>
              </div>
              <div class="code-content" v-if="state.testResult?.simplifiedCode">
                <pre class="code-block">{{ state.testResult.simplifiedCode }}</pre>
              </div>
              <div v-else class="placeholder-content">
                完成字段分析后将显示简化代码
              </div>
            </div>

          </div>
        </div>
      </div>




    </div>
  `,

  setup() {
    // 响应式数据
    const state = reactive({
      // 输入相关
      inputText: '',
      inputType: 'auto',
      detectedType: '',
      isParsing: false,

      // 解析结果
      parsedRequest: null,
      pythonCode: '',
      requestSummary: {},

      // 验证类型选择
      validationType: 'textMatching', // 'textMatching', 'lengthRange'

      // 测试配置
      validationConfig: {
        expression: '',
        timeout: 30,
        maxRetries: 3,

        followRedirect: true,
        userAgent: 'RequestProbe/1.0',
        textMatching: {
          enabled: true, // 默认启用文本匹配验证
          texts: [],
          matchMode: 'all', // 默认改为全部匹配
          caseSensitive: false
        },
        lengthRange: {
          enabled: false,
          minLength: 0,
          maxLength: -1
        },
        useCustomExpr: false,
        encodingConfig: {
          enabled: false,
          calibrationText: '',
          supportedEncodings: ['UTF-8', 'GBK', 'GB2312', 'Big5'],
          detectedEncoding: 'UTF-8'
        }
      },

      // 辅助输入
      textMatchingInput: '',



      // 测试结果
      testResult: null,
      testStatistics: {},
      isTestRunning: false,


      // 单次请求测试结果
      singleTestResult: null,
      isSingleTesting: false,

      // UI状态
      activeTab: 'python',
      leftPanelWidth: '45%',

      // 演示功能（保持向后兼容）
      name: '',
      greeting: '',
      requestInput: '',
      parseResult: null
    });

    // 计算属性
    const canTest = computed(() => {
      return state.parsedRequest && !state.isTestRunning;
    });

    const testResultSummary = computed(() => {
      if (!state.testResult) return null;

      const requiredHeaders = state.testResult.headerResults?.filter(r => r.isRequired).length || 0;
      const requiredCookies = state.testResult.cookieResults?.filter(r => r.isRequired).length || 0;
      const totalHeaders = state.testResult.headerResults?.length || 0;
      const totalCookies = state.testResult.cookieResults?.length || 0;

      return {
        requiredHeaders,
        requiredCookies,
        totalHeaders,
        totalCookies,
        originalPassed: state.testResult.originalPassed,
        testDuration: state.testResult.testDuration
      };
    });

    const headerTestResults = computed(() => {
      if (!state.testResult || !state.testResult.headerResults) return [];
      return state.testResult.headerResults;
    });

    const cookieTestResults = computed(() => {
      if (!state.testResult || !state.testResult.cookieResults) return [];
      return state.testResult.cookieResults;
    });

    const allTestResults = computed(() => {
      if (!state.testResult) return [];

      const results = [];
      if (state.testResult.headerResults) {
        results.push(...state.testResult.headerResults);
      }
      if (state.testResult.cookieResults) {
        results.push(...state.testResult.cookieResults);
      }
      return results;
    });

    // 方法
    const detectInputType = async () => {
      if (!state.inputText.trim()) {
        state.detectedType = '';
        return;
      }

      try {
        const type = await DetectInputType(state.inputText);
        state.detectedType = type;
      } catch (error) {
        console.error('检测输入类型失败:', error);
      }
    };

    const parseRequest = async () => {
      if (!state.inputText.trim()) {
        ElMessage.warning('请输入HTTP请求或Curl命令');
        return;
      }

      state.isParsing = true;
      try {
        let request;
        if (state.inputType === 'auto') {
          request = await ParseRequest(state.inputText);
        } else {
          request = await ParseRequestWithType(state.inputText, state.inputType);
        }

        state.parsedRequest = request;

        // 生成Python代码
        const code = await GeneratePythonCode(request);
        state.pythonCode = code;

        // 获取请求摘要
        const summary = await GetRequestSummary(request);
        state.requestSummary = summary;

        ElMessage.success('请求解析成功');

        // 自动进行测试请求
        await testRequestOnly();

      } catch (error) {
        const errorMessage = error?.message || error?.toString() || '未知错误';
        ElMessage.error('解析失败: ' + errorMessage);
        console.error('解析请求失败:', error);
      } finally {
        state.isParsing = false;
      }
    };

    const testRequestOnly = async () => {
      if (!state.parsedRequest) {
        ElMessage.warning('请先解析请求');
        return;
      }

      state.isSingleTesting = true;
      // 重置状态
      state.singleTestResult = null;
      state.detectedEncoding = '';
      state.displayBody = '';
      state.encodingVerified = false;

      try {
        // 转换超时时间为纳秒（Go的time.Duration）
        const config = {
          ...state.validationConfig,
          timeout: state.validationConfig.timeout * 1000000000 // 秒转纳秒
        };

        const result = await TestRequestOnly(state.parsedRequest, config);
        state.singleTestResult = result;

        ElMessage.success('请求测试完成，请检查编码并确认');
        state.activeTab = 'request-test';



      } catch (error) {
        const errorMessage = error?.message || error?.toString() || '未知错误';
        ElMessage.error('请求测试失败: ' + errorMessage);
        console.error('请求测试失败:', error);
      } finally {
        state.isSingleTesting = false;
      }
    };



    const testFieldNecessity = async () => {
      if (!state.parsedRequest) {
        ElMessage.warning('请先解析请求');
        return;
      }

      state.isTestRunning = true;
      state.testProgress = { currentStep: '准备测试...', progress: 0, message: '准备测试...' };

      const loading = ElLoading.service({
        lock: true,
        text: '正在测试字段必要性...',
        spinner: 'el-icon-loading',
        background: 'rgba(0, 0, 0, 0.7)'
      });

      try {
        // 转换超时时间为纳秒（Go的time.Duration）
        const config = {
          ...state.validationConfig,
          timeout: state.validationConfig.timeout * 1000000000 // 秒转纳秒
        };

        const result = await TestFieldNecessity(state.parsedRequest, config);

        state.testResult = result;

        // 获取测试统计
        const stats = await GetTestStatistics(result);
        state.testStatistics = stats;

        ElMessage.success('字段分析完成');
        state.activeTab = 'result';

      } catch (error) {
        const errorMessage = error?.message || error?.toString() || '未知错误';
        ElMessage.error('测试失败: ' + errorMessage);
        console.error('测试失败:', error);
      } finally {
        state.isTestRunning = false;
        loading.close();
      }
    };





    const loadDefaultConfig = async () => {
      try {
        const config = await GetDefaultValidationConfig();
        // 转换纳秒为秒
        config.timeout = config.timeout / 1000000000;
        state.validationConfig = config;
      } catch (error) {
        console.error('加载默认配置失败:', error);
      }
    };

    const copyCode = async (code) => {
      try {
        await navigator.clipboard.writeText(code);
        ElMessage.success('代码已复制到剪贴板');
      } catch (error) {
        ElMessage.error('复制失败');
      }
    };

    const formatDuration = (duration) => {
      if (!duration) return '0ms';
      // duration是纳秒，转换为毫秒
      const ms = duration / 1000000;
      if (ms < 1000) {
        return `${ms.toFixed(0)}ms`;
      } else {
        return `${(ms / 1000).toFixed(2)}s`;
      }
    };

    const formatBytes = (bytes) => {
      if (bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const getStatusCodeType = (statusCode) => {
      if (statusCode >= 200 && statusCode < 300) return 'success';
      if (statusCode >= 300 && statusCode < 400) return 'warning';
      if (statusCode >= 400) return 'danger';
      return 'info';
    };

    const formatHeaders = (headers) => {
      if (!headers) return [];
      return Object.entries(headers).map(([name, value]) => ({ name, value }));
    };

    const updateTextMatching = () => {
      const texts = state.textMatchingInput
        .split('\n')
        .map(text => text.trim())
        .filter(text => text.length > 0);
      state.validationConfig.textMatching.texts = texts;
    };

    // 验证类型变化处理
    const onValidationTypeChange = (type) => {
      // 重置所有验证配置
      state.validationConfig.textMatching.enabled = false;
      state.validationConfig.lengthRange.enabled = false;

      // 根据选择的类型启用对应的验证
      if (type === 'textMatching') {
        state.validationConfig.textMatching.enabled = true;
      } else if (type === 'lengthRange') {
        state.validationConfig.lengthRange.enabled = true;
      }
    };





    const clearResults = () => {
      // 清空测试结果
      state.singleTestResult = null;
      state.testResult = null;
      state.testStatistics = {};

      // 清空输入内容
      state.inputText = '';
      state.detectedType = '';

      // 清空解析结果
      state.parsedRequest = null;
      state.pythonCode = '';
      state.requestSummary = {};

      // 重置验证配置输入
      state.textMatchingInput = '';
      state.validationConfig.textMatching.texts = [];

      // 重置到默认标签页
      state.activeTab = 'python';

      ElMessage.success('所有内容已清空');
    };

    // 演示功能（保持向后兼容）
    const greetUser = async () => {
      if (!state.name.trim()) {
        ElMessage.warning('请输入您的名字');
        return;
      }

      try {
        const greeting = await Greet(state.name);
        state.greeting = greeting;
        ElMessage.success('问候成功！');
      } catch (error) {
        const errorMessage = error?.message || error?.toString() || '未知错误';
        ElMessage.error('问候失败: ' + errorMessage);
        console.error('问候失败:', error);
      }
    };

    const parseSimpleRequest = async () => {
      if (!state.requestInput.trim()) {
        ElMessage.warning('请输入一些文本');
        return;
      }

      try {
        const result = await ParseSimpleRequest(state.requestInput);
        state.parseResult = result;
        ElMessage.success('解析成功！');
      } catch (error) {
        const errorMessage = error?.message || error?.toString() || '未知错误';
        ElMessage.error('解析失败: ' + errorMessage);
        console.error('解析失败:', error);
      }
    };

    // 生命周期
    onMounted(async () => {
      await loadDefaultConfig();
      await loadEncodings();

      // 确保默认选择的验证类型被正确启用
      onValidationTypeChange(state.validationType);

      // 初始化文本匹配输入
      state.textMatchingInput = state.validationConfig.textMatching.texts.join('\n');
    });

    return {
      state,
      canTest,
      testResultSummary,
      headerTestResults,
      cookieTestResults,
      allTestResults,
      detectInputType,
      parseRequest,
      testRequestOnly,
      testFieldNecessity,

      loadDefaultConfig,
      copyCode,
      formatDuration,
      formatBytes,
      getStatusCodeType,
      formatHeaders,
      updateTextMatching,
      onValidationTypeChange,
      clearResults,
      greetUser,
      parseSimpleRequest
    };
  }
});

// 使用Element Plus
app.use(ElementPlus);

// 挂载应用
app.mount('#app');
