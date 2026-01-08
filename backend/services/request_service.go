package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"RequestProbe/backend/core/manager"
	"RequestProbe/backend/core/parser"
	"RequestProbe/backend/core/tester"
	"RequestProbe/backend/models"
)

// RequestService 请求服务
type RequestService struct {
	parser            *parser.UnifiedRequestParser
	tester            *tester.RequestTester
	expressionManager *manager.ExpressionManager
}

// NewRequestService 创建请求服务
func NewRequestService() *RequestService {
	return &RequestService{
		parser:            parser.NewUnifiedRequestParser(),
		tester:            tester.NewRequestTester(),
		expressionManager: manager.NewExpressionManager(),
	}
}

// ParseRequest 解析请求
func (s *RequestService) ParseRequest(ctx context.Context, input string) (*models.ParsedRequest, error) {
	request, err := s.parser.Parse(input)
	if err != nil {
		return nil, err
	}

	// 验证请求
	if err := s.parser.ValidateRequest(request); err != nil {
		return nil, err
	}

	return request, nil
}

// ParseRequestWithType 使用指定类型解析请求
func (s *RequestService) ParseRequestWithType(ctx context.Context, input, inputType string) (*models.ParsedRequest, error) {
	request, err := s.parser.ParseWithType(input, inputType)
	if err != nil {
		return nil, err
	}

	// 验证请求
	if err := s.parser.ValidateRequest(request); err != nil {
		return nil, err
	}

	return request, nil
}

// DetectInputType 检测输入类型
func (s *RequestService) DetectInputType(ctx context.Context, input string) string {
	return s.parser.DetectInputType(input)
}

// GeneratePythonCode 生成Python代码
func (s *RequestService) GeneratePythonCode(ctx context.Context, request *models.ParsedRequest) string {
	return s.parser.GeneratePythonCode(request)
}

// TestSingleRequest 测试单个请求
func (s *RequestService) TestSingleRequest(ctx context.Context, request *models.ParsedRequest, config *models.ValidationConfig) (*models.ResponseData, error) {
	// 设置超时
	if config.Timeout > 0 {
		s.tester.SetTimeout(config.Timeout)
	}

	return s.tester.TestRequest(request, config)
}

// TestFieldNecessity 测试字段必要性
func (s *RequestService) TestFieldNecessity(ctx context.Context, request *models.ParsedRequest, config *models.ValidationConfig, progressCallback func(*models.TestProgress)) (*models.BatchTestResult, error) {
	// 设置超时
	if config.Timeout > 0 {
		s.tester.SetTimeout(config.Timeout)
	} else {
		s.tester.SetTimeout(30 * time.Second) // 默认30秒超时
	}

	return s.tester.BatchTestFieldNecessity(request, config, progressCallback)
}

// ValidateExpression 验证表达式
func (s *RequestService) ValidateExpression(ctx context.Context, expression string) error {
	return s.expressionManager.ValidateExpression(expression)
}

// GetExpressionTemplates 获取表达式模板
func (s *RequestService) GetExpressionTemplates(ctx context.Context) []models.ExpressionTemplate {
	return s.expressionManager.GetAllTemplates()
}

// GetExpressionTemplatesByCategory 按分类获取模板
func (s *RequestService) GetExpressionTemplatesByCategory(ctx context.Context, category string) []models.ExpressionTemplate {
	return s.expressionManager.GetTemplatesByCategory(category)
}

// GetExpressionCategories 获取表达式分类
func (s *RequestService) GetExpressionCategories(ctx context.Context) []string {
	return s.expressionManager.GetCategories()
}

// AddExpressionTemplate 添加表达式模板
func (s *RequestService) AddExpressionTemplate(ctx context.Context, template models.ExpressionTemplate) error {
	return s.expressionManager.AddTemplate(template)
}

// UpdateExpressionTemplate 更新表达式模板
func (s *RequestService) UpdateExpressionTemplate(ctx context.Context, template models.ExpressionTemplate) error {
	return s.expressionManager.UpdateTemplate(template)
}

// DeleteExpressionTemplate 删除表达式模板
func (s *RequestService) DeleteExpressionTemplate(ctx context.Context, id string) error {
	return s.expressionManager.DeleteTemplate(id)
}

// ExportExpressionTemplates 导出表达式模板
func (s *RequestService) ExportExpressionTemplates(ctx context.Context, filePath string) error {
	return s.expressionManager.ExportTemplates(filePath)
}

// ImportExpressionTemplates 导入表达式模板
func (s *RequestService) ImportExpressionTemplates(ctx context.Context, filePath string) error {
	return s.expressionManager.ImportTemplates(filePath)
}

// GetDefaultValidationConfig 获取默认验证配置
func (s *RequestService) GetDefaultValidationConfig(ctx context.Context) *models.ValidationConfig {
	return &models.ValidationConfig{
		Expression: "", // 不再使用表达式
		Timeout:    30 * time.Second,
		MaxRetries: 3, // 默认重试3次

		FollowRedirect: true,
		UserAgent:      "RequestProbe/1.0",

		// 新的验证配置
		TextMatching: models.TextMatchingConfig{
			Enabled:       true,       // 默认启用，与前端保持一致
			Texts:         []string{}, // 默认为空，用户可以添加
			MatchMode:     "all",      // 默认全部匹配，与前端保持一致
			CaseSensitive: false,      // 默认不区分大小写，与前端保持一致
		},
		LengthRange: models.LengthRangeConfig{
			Enabled:   false, // 默认关闭
			MinLength: 0,     // 默认最小长度0
			MaxLength: -1,    // 默认无最大长度限制
		},
		UseCustomExpr: false, // 默认不使用自定义表达式

		// 编码配置
		EncodingConfig: models.EncodingConfig{
			Enabled:            false,                                      // 默认关闭编码检测
			CalibrationText:    "",                                         // 默认无校准文本
			SupportedEncodings: []string{"UTF-8", "GBK", "GB2312", "Big5"}, // 常用编码
			DetectedEncoding:   "UTF-8",                                    // 默认UTF-8
		},

		// 字段保留配置
		PreserveUserAgent: true, // 默认保留User-Agent
	}
}

// TestRequestWithRetry 带重试的请求测试
func (s *RequestService) TestRequestWithRetry(ctx context.Context, request *models.ParsedRequest, config *models.ValidationConfig) (*models.ResponseData, error) {
	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		response, err := s.TestSingleRequest(ctx, request, config)
		if err == nil {
			return response, nil
		}
		lastErr = err

		// 如果不是最后一次重试，等待一段时间
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Second * time.Duration(i+1)):
				// 指数退避
			}
		}
	}

	return nil, fmt.Errorf("重试 %d 次后仍然失败: %v", maxRetries, lastErr)
}

// GetRequestSummary 获取请求摘要信息
func (s *RequestService) GetRequestSummary(ctx context.Context, request *models.ParsedRequest) map[string]interface{} {
	summary := map[string]interface{}{
		"method":      request.Method,
		"url":         request.URL,
		"headerCount": len(request.Headers),
		"cookieCount": len(request.Cookies),
		"hasBody":     request.Body != "",
		"contentType": request.ContentType,
		"queryParams": len(request.QueryParams),
	}

	// 分析请求类型
	if request.ContentType != "" {
		if strings.Contains(strings.ToLower(request.ContentType), "json") {
			summary["requestType"] = "JSON"
		} else if strings.Contains(strings.ToLower(request.ContentType), "form") {
			summary["requestType"] = "Form"
		} else {
			summary["requestType"] = "Other"
		}
	} else {
		summary["requestType"] = "Unknown"
	}

	return summary
}

// GetTestStatistics 获取测试统计信息
func (s *RequestService) GetTestStatistics(ctx context.Context, result *models.BatchTestResult) map[string]interface{} {
	stats := map[string]interface{}{
		"totalTests":     result.TotalTests,
		"passedTests":    result.PassedTests,
		"failedTests":    result.TotalTests - result.PassedTests,
		"testDuration":   result.TestDuration.String(),
		"originalPassed": result.OriginalPassed,
	}

	// 计算必需字段统计
	requiredHeaders := 0
	optionalHeaders := 0
	for _, headerResult := range result.HeaderResults {
		if headerResult.IsRequired {
			requiredHeaders++
		} else {
			optionalHeaders++
		}
	}

	requiredCookies := 0
	optionalCookies := 0
	for _, cookieResult := range result.CookieResults {
		if cookieResult.IsRequired {
			requiredCookies++
		} else {
			optionalCookies++
		}
	}

	stats["requiredHeaders"] = requiredHeaders
	stats["optionalHeaders"] = optionalHeaders
	stats["requiredCookies"] = requiredCookies
	stats["optionalCookies"] = optionalCookies

	// 计算简化率
	originalFieldCount := len(result.OriginalRequest.Headers) + len(result.OriginalRequest.Cookies)
	simplifiedFieldCount := len(result.SimplifiedRequest.Headers) + len(result.SimplifiedRequest.Cookies)

	if originalFieldCount > 0 {
		simplificationRate := float64(originalFieldCount-simplifiedFieldCount) / float64(originalFieldCount) * 100
		stats["simplificationRate"] = fmt.Sprintf("%.1f%%", simplificationRate)
	} else {
		stats["simplificationRate"] = "0%"
	}

	return stats
}

// GetSupportedEncodings 获取支持的编码列表
func (s *RequestService) GetSupportedEncodings(ctx context.Context) []string {
	return s.tester.Validator.GetSupportedEncodings()
}

// GetCommonEncodings 获取常用编码列表
func (s *RequestService) GetCommonEncodings(ctx context.Context) []string {
	return s.tester.Validator.GetCommonEncodings()
}

// DetectEncoding 检测响应编码
func (s *RequestService) DetectEncoding(ctx context.Context, responseBody []byte, calibrationText string) (string, error) {
	return s.tester.Validator.DetectEncoding(responseBody, calibrationText)
}

// DecodeResponse 使用指定编码解码响应
func (s *RequestService) DecodeResponse(ctx context.Context, responseBody []byte, encodingName string) (string, error) {
	return s.tester.Validator.DecodeResponse(responseBody, encodingName)
}

// DetectEncodingFromResponse 从响应数据中检测编码
func (s *RequestService) DetectEncodingFromResponse(ctx context.Context, response *models.ResponseData, calibrationText string) (string, error) {
	if response.RawBody == nil {
		// 如果没有原始字节数据，使用字符串转换
		return s.tester.Validator.DetectEncoding([]byte(response.Body), calibrationText)
	}
	return s.tester.Validator.DetectEncoding(response.RawBody, calibrationText)
}

// DecodeResponseFromResponse 从响应数据中解码
func (s *RequestService) DecodeResponseFromResponse(ctx context.Context, response *models.ResponseData, encodingName string) (string, error) {
	if response.RawBody == nil {
		// 如果没有原始字节数据，使用字符串转换
		return s.tester.Validator.DecodeResponse([]byte(response.Body), encodingName)
	}
	return s.tester.Validator.DecodeResponse(response.RawBody, encodingName)
}

// AutoDetectEncodingFromResponse 自动检测响应编码并转换
func (s *RequestService) AutoDetectEncodingFromResponse(ctx context.Context, response *models.ResponseData) (string, string, error) {
	if response.RawBody == nil {
		// 如果没有原始字节数据，使用字符串转换
		return s.tester.Validator.AutoDetectEncoding([]byte(response.Body))
	}
	return s.tester.Validator.AutoDetectEncoding(response.RawBody)
}
