package parser

import (
	"fmt"
	"strings"

	"RequestProbe/backend/models"
)

// RequestParser 请求解析器接口
type RequestParser interface {
	Parse(input string) (*models.ParsedRequest, error)
}

// UnifiedRequestParser 统一请求解析器
type UnifiedRequestParser struct {
	rawParser  *RawRequestParser
	curlParser *CurlRequestParser
}

// NewUnifiedRequestParser 创建统一解析器
func NewUnifiedRequestParser() *UnifiedRequestParser {
	return &UnifiedRequestParser{
		rawParser:  NewRawRequestParser(),
		curlParser: NewCurlRequestParser(),
	}
}

// Parse 自动检测格式并解析请求
func (p *UnifiedRequestParser) Parse(input string) (*models.ParsedRequest, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("输入内容不能为空")
	}

	// 检测输入格式
	inputType := p.DetectInputType(input)

	switch inputType {
	case "curl":
		return p.curlParser.Parse(input)
	case "raw":
		return p.rawParser.Parse(input)
	default:
		return nil, fmt.Errorf("无法识别的请求格式，请使用Raw HTTP格式或Curl命令")
	}
}

// DetectInputType 检测输入类型
func (p *UnifiedRequestParser) DetectInputType(input string) string {
	trimmed := strings.TrimSpace(input)

	// 检测是否为Curl命令
	if p.curlParser.IsCurlCommand(trimmed) {
		return "curl"
	}

	// 检测是否为Raw HTTP请求
	if p.rawParser.IsRawRequest(trimmed) {
		return "raw"
	}

	return "unknown"
}

// ParseWithType 使用指定类型解析请求
func (p *UnifiedRequestParser) ParseWithType(input, inputType string) (*models.ParsedRequest, error) {
	switch strings.ToLower(inputType) {
	case "curl":
		return p.curlParser.Parse(input)
	case "raw", "http":
		return p.rawParser.Parse(input)
	default:
		return nil, fmt.Errorf("不支持的输入类型: %s", inputType)
	}
}

// ValidateRequest 验证解析后的请求
func (p *UnifiedRequestParser) ValidateRequest(req *models.ParsedRequest) error {
	if req == nil {
		return fmt.Errorf("请求对象不能为空")
	}

	if req.Method == "" {
		return fmt.Errorf("HTTP方法不能为空")
	}

	if req.URL == "" {
		return fmt.Errorf("请求URL不能为空")
	}

	// 验证URL格式
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		return fmt.Errorf("URL必须以http://或https://开头")
	}

	// 验证HTTP方法
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	isValidMethod := false
	for _, method := range validMethods {
		if req.Method == method {
			isValidMethod = true
			break
		}
	}

	if !isValidMethod {
		return fmt.Errorf("不支持的HTTP方法: %s", req.Method)
	}

	return nil
}

// GeneratePythonCode 生成Python requests代码
func (p *UnifiedRequestParser) GeneratePythonCode(req *models.ParsedRequest) string {
	var code strings.Builder

	code.WriteString("import requests\n\n")

	// Headers
	if len(req.Headers) > 0 {
		code.WriteString("headers = {\n")
		for key, value := range req.Headers {
			// 跳过Cookie header，因为会单独处理
			if strings.ToLower(key) != "cookie" {
				code.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", key, value))
			}
		}
		code.WriteString("}\n")
	}

	// Cookies
	if len(req.Cookies) > 0 {
		code.WriteString("cookies = {\n")
		for key, value := range req.Cookies {
			code.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", key, value))
		}
		code.WriteString("}\n")
	}

	// 解析URL和参数
	baseURL, queryParams := p.parseURLAndParams(req.URL)
	code.WriteString(fmt.Sprintf("url = \"%s\"\n", baseURL))

	// 查询参数
	if len(queryParams) > 0 {
		code.WriteString("params = {\n")
		for key, value := range queryParams {
			code.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", key, value))
		}
		code.WriteString("}\n")
	}

	// 请求体
	var dataParam string
	if req.Body != "" {
		// 尝试判断是否为JSON
		if strings.HasPrefix(strings.TrimSpace(req.Body), "{") || strings.HasPrefix(strings.TrimSpace(req.Body), "[") {
			code.WriteString(fmt.Sprintf("data = %s\n", req.Body))
			dataParam = "json=data"
		} else {
			code.WriteString(fmt.Sprintf("data = \"%s\"\n", req.Body))
			dataParam = "data=data"
		}
	}

	// 构建请求调用
	code.WriteString(fmt.Sprintf("response = requests.%s(url", strings.ToLower(req.Method)))

	if len(req.Headers) > 0 {
		code.WriteString(", headers=headers")
	}

	if len(req.Cookies) > 0 {
		code.WriteString(", cookies=cookies")
	}

	if len(queryParams) > 0 {
		code.WriteString(", params=params")
	}

	if dataParam != "" {
		code.WriteString(fmt.Sprintf(", %s", dataParam))
	}

	code.WriteString(")\n\n")
	code.WriteString("print(response.text)\n")
	code.WriteString("print(response)")

	return code.String()
}

// parseURLAndParams 解析URL，分离基础URL和查询参数
func (p *UnifiedRequestParser) parseURLAndParams(fullURL string) (string, map[string]string) {
	parts := strings.Split(fullURL, "?")
	baseURL := parts[0]
	queryParams := make(map[string]string)

	if len(parts) > 1 {
		// 解析查询参数
		paramPairs := strings.Split(parts[1], "&")
		for _, pair := range paramPairs {
			if keyValue := strings.Split(pair, "="); len(keyValue) == 2 {
				queryParams[keyValue[0]] = keyValue[1]
			}
		}
	}

	return baseURL, queryParams
}
