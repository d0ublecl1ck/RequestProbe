package models

import "time"

// ParsedRequest 表示解析后的HTTP请求
type ParsedRequest struct {
	Method      string            `json:"method"`      // HTTP方法
	URL         string            `json:"url"`         // 请求URL
	Headers     map[string]string `json:"headers"`     // 请求头
	Cookies     map[string]string `json:"cookies"`     // Cookie字段
	Body        string            `json:"body"`        // 请求体
	QueryParams map[string]string `json:"queryParams"` // URL查询参数
	ContentType string            `json:"contentType"` // 内容类型
}

// CumulativeTestState 累积测试状态
type CumulativeTestState struct {
	Headers map[string]string `json:"headers"` // 当前有效的Headers
	Cookies map[string]string `json:"cookies"` // 当前有效的Cookies
}

// DeepCopy 深拷贝累积测试状态
func (s *CumulativeTestState) DeepCopy() *CumulativeTestState {
	newState := &CumulativeTestState{
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	for k, v := range s.Headers {
		newState.Headers[k] = v
	}
	for k, v := range s.Cookies {
		newState.Cookies[k] = v
	}

	return newState
}

// FieldTestResult 单个字段的测试结果（累积模式）
type FieldTestResult struct {
	Required   bool                 `json:"required"`   // 是否必需
	Value      string               `json:"value"`      // 字段值
	TestResult *SingleRequestResult `json:"testResult"` // 测试结果详情
}

// SingleRequestResult 单次请求结果
type SingleRequestResult struct {
	Success      bool          `json:"success"`      // 是否成功
	Error        string        `json:"error"`        // 错误信息
	Note         string        `json:"note"`         // 备注信息
	ResponseInfo *ResponseInfo `json:"responseInfo"` // 响应信息
}

// ResponseInfo 响应信息
type ResponseInfo struct {
	StatusCode int               `json:"statusCode"` // 状态码
	URL        string            `json:"url"`        // URL
	Headers    map[string]string `json:"headers"`    // 响应头
}

// TestResults 累积测试结果
type TestResults struct {
	Headers map[string]*FieldTestResult `json:"headers"` // Header测试结果
	Cookies map[string]*FieldTestResult `json:"cookies"` // Cookie测试结果
}

// TestResult 表示单个字段的测试结果（保持向后兼容）
type TestResult struct {
	FieldName   string `json:"fieldName"`   // 字段名称
	FieldType   string `json:"fieldType"`   // 字段类型 (header/cookie)
	IsRequired  bool   `json:"isRequired"`  // 是否必需
	TestPassed  bool   `json:"testPassed"`  // 测试是否通过
	ErrorMsg    string `json:"errorMsg"`    // 错误信息
	StatusCode  int    `json:"statusCode"`  // 响应状态码
	ResponseMsg string `json:"responseMsg"` // 响应消息
}

// BatchTestResult 表示批量测试结果
type BatchTestResult struct {
	OriginalRequest   *ParsedRequest `json:"originalRequest"`   // 原始请求
	OriginalPassed    bool           `json:"originalPassed"`    // 原始请求是否通过
	OriginalError     string         `json:"originalError"`     // 原始请求错误
	HeaderResults     []TestResult   `json:"headerResults"`     // Header测试结果
	CookieResults     []TestResult   `json:"cookieResults"`     // Cookie测试结果
	SimplifiedRequest *ParsedRequest `json:"simplifiedRequest"` // 简化后的请求
	SimplifiedCode    string         `json:"simplifiedCode"`    // 简化后的Python代码
	TestDuration      time.Duration  `json:"testDuration"`      // 测试耗时
	TotalTests        int            `json:"totalTests"`        // 总测试数
	PassedTests       int            `json:"passedTests"`       // 通过测试数

	// 新增累积测试结果
	CumulativeResults *TestResults `json:"cumulativeResults"` // 累积测试结果
}

// ValidationConfig 表示验证配置
type ValidationConfig struct {
	Expression string        `json:"expression"` // 验证表达式（已弃用，保持兼容性）
	Timeout    time.Duration `json:"timeout"`    // 请求超时时间
	MaxRetries int           `json:"maxRetries"` // 最大重试次数

	FollowRedirect bool   `json:"followRedirect"` // 是否跟随重定向
	UserAgent      string `json:"userAgent"`      // User-Agent

	// 新的验证配置
	TextMatching  TextMatchingConfig `json:"textMatching"`  // 文本匹配配置
	LengthRange   LengthRangeConfig  `json:"lengthRange"`   // 长度范围配置
	UseCustomExpr bool               `json:"useCustomExpr"` // 是否使用自定义表达式

	// 编码配置
	EncodingConfig EncodingConfig `json:"encodingConfig"` // 编码配置

	// 字段保留配置
	PreserveUserAgent bool `json:"preserveUserAgent"` // 默认保留User-Agent（无论测试结果如何）
}

// TextMatchingConfig 文本匹配配置
type TextMatchingConfig struct {
	Enabled       bool     `json:"enabled"`       // 是否启用文本匹配
	Texts         []string `json:"texts"`         // 要匹配的文本列表
	MatchMode     string   `json:"matchMode"`     // 匹配模式：all（全部匹配）或 any（任意匹配）
	CaseSensitive bool     `json:"caseSensitive"` // 是否区分大小写
}

// LengthRangeConfig 长度范围配置
type LengthRangeConfig struct {
	Enabled   bool `json:"enabled"`   // 是否启用长度检查
	MinLength int  `json:"minLength"` // 最小长度
	MaxLength int  `json:"maxLength"` // 最大长度（-1表示无限制）
}

// EncodingConfig 编码配置
type EncodingConfig struct {
	Enabled            bool     `json:"enabled"`            // 是否启用编码检测
	CalibrationText    string   `json:"calibrationText"`    // 校准文本
	SupportedEncodings []string `json:"supportedEncodings"` // 支持的编码列表
	DetectedEncoding   string   `json:"detectedEncoding"`   // 检测到的编码
}

// ResponseData 表示HTTP响应数据
type ResponseData struct {
	StatusCode       int               `json:"statusCode"`       // 状态码
	Headers          map[string]string `json:"headers"`          // 响应头
	Body             string            `json:"body"`             // 响应体
	Cookies          []ResponseCookie  `json:"cookies"`          // 响应Cookie
	URL              string            `json:"url"`              // 最终URL
	Duration         time.Duration     `json:"duration"`         // 请求耗时
	ContentLength    int64             `json:"contentLength"`    // 响应大小（字节）
	CharacterCount   int               `json:"characterCount"`   // 响应字符长度
	RawBody          []byte            `json:"-"`                // 原始响应字节（不序列化到JSON）
	DetectedEncoding string            `json:"detectedEncoding"` // 检测到的编码
}

// ResponseCookie 表示响应 Cookie（避免暴露 time.Time）
type ResponseCookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

// ExpressionTemplate 表示验证表达式模板
type ExpressionTemplate struct {
	ID          string `json:"id"`          // 模板ID
	Name        string `json:"name"`        // 模板名称
	Description string `json:"description"` // 模板描述
	Expression  string `json:"expression"`  // 表达式内容
	Category    string `json:"category"`    // 分类
}

// TestProgress 表示测试进度
type TestProgress struct {
	CurrentStep    string      `json:"currentStep"`    // 当前步骤
	TotalSteps     int         `json:"totalSteps"`     // 总步骤数
	CompletedSteps int         `json:"completedSteps"` // 已完成步骤数
	Progress       float64     `json:"progress"`       // 进度百分比
	Message        string      `json:"message"`        // 进度消息
	FieldResult    *TestResult `json:"fieldResult"`    // 单个字段的测试结果（可选）
}
