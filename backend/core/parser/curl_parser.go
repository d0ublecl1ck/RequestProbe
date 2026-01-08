package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"RequestProbe/backend/models"
)

// CurlRequestParser Curl命令解析器
type CurlRequestParser struct{}

// NewCurlRequestParser 创建Curl解析器
func NewCurlRequestParser() *CurlRequestParser {
	return &CurlRequestParser{}
}

// Parse 解析Curl命令
func (p *CurlRequestParser) Parse(curlCommand string) (*models.ParsedRequest, error) {
	if strings.TrimSpace(curlCommand) == "" {
		return nil, fmt.Errorf("Curl命令不能为空")
	}

	// 清理命令，处理多行情况
	cleanCommand := p.cleanCurlCommand(curlCommand)

	// 解析命令参数
	args, err := p.parseCurlArgs(cleanCommand)
	if err != nil {
		return nil, fmt.Errorf("解析Curl参数失败: %v", err)
	}

	// 提取URL
	requestURL := p.extractURL(args)
	if requestURL == "" {
		return nil, fmt.Errorf("未找到请求URL")
	}

	// 提取HTTP方法
	method := p.extractMethod(args)

	// 提取Headers
	headers := p.extractHeaders(args)

	// 提取Cookies
	cookies := p.extractCookies(args)

	// 提取请求体
	body := p.extractBody(args)

	// 解析URL参数
	queryParams, err := p.parseQueryParams(requestURL)
	if err != nil {
		return nil, fmt.Errorf("解析URL参数失败: %v", err)
	}

	// 确定Content-Type
	contentType := headers["Content-Type"]
	if contentType == "" {
		contentType = headers["content-type"]
	}

	return &models.ParsedRequest{
		Method:      method,
		URL:         requestURL,
		Headers:     headers,
		Cookies:     cookies,
		Body:        body,
		QueryParams: queryParams,
		ContentType: contentType,
	}, nil
}

// cleanCurlCommand 清理Curl命令，处理多行和转义
func (p *CurlRequestParser) cleanCurlCommand(command string) string {
	// 移除行尾的反斜杠和换行符
	command = regexp.MustCompile(`\\\s*\n\s*`).ReplaceAllString(command, " ")

	// 移除多余的空白字符
	command = regexp.MustCompile(`\s+`).ReplaceAllString(command, " ")

	return strings.TrimSpace(command)
}

// parseCurlArgs 解析Curl命令参数
func (p *CurlRequestParser) parseCurlArgs(command string) ([]string, error) {
	var args []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune
	var escaped bool

	for _, char := range command {
		if escaped {
			current.WriteRune(char)
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if !inQuotes && (char == '"' || char == '\'') {
			inQuotes = true
			quoteChar = char
			continue
		}

		if inQuotes && char == quoteChar {
			inQuotes = false
			continue
		}

		if !inQuotes && char == ' ' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(char)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

// extractURL 提取URL
func (p *CurlRequestParser) extractURL(args []string) string {
	for i, arg := range args {
		if arg == "curl" {
			continue
		}
		if strings.HasPrefix(arg, "-") {
			// 跳过选项参数
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++ // 跳过选项的值
			}
			continue
		}
		// 第一个非选项参数应该是URL
		return arg
	}
	return ""
}

// extractMethod 提取HTTP方法
func (p *CurlRequestParser) extractMethod(args []string) string {
	for i, arg := range args {
		if (arg == "-X" || arg == "--request") && i+1 < len(args) {
			return strings.ToUpper(args[i+1])
		}
	}
	return "GET" // 默认方法
}

// extractHeaders 提取Headers
func (p *CurlRequestParser) extractHeaders(args []string) map[string]string {
	headers := make(map[string]string)

	for i, arg := range args {
		if (arg == "-H" || arg == "--header") && i+1 < len(args) {
			headerValue := args[i+1]
			if colonIndex := strings.Index(headerValue, ":"); colonIndex > 0 {
				key := strings.TrimSpace(headerValue[:colonIndex])
				value := strings.TrimSpace(headerValue[colonIndex+1:])
				headers[key] = value
			}
		}
	}

	return headers
}

// extractCookies 提取Cookies
func (p *CurlRequestParser) extractCookies(args []string) map[string]string {
	cookies := make(map[string]string)

	for i, arg := range args {
		if (arg == "-b" || arg == "--cookie") && i+1 < len(args) {
			cookieValue := args[i+1]

			// 解析cookie字符串
			pairs := strings.Split(cookieValue, ";")
			for _, pair := range pairs {
				pair = strings.TrimSpace(pair)
				if equalIndex := strings.Index(pair, "="); equalIndex > 0 {
					name := strings.TrimSpace(pair[:equalIndex])
					value := strings.TrimSpace(pair[equalIndex+1:])
					cookies[name] = value
				}
			}
		}
	}

	return cookies
}

// extractBody 提取请求体
func (p *CurlRequestParser) extractBody(args []string) string {
	for i, arg := range args {
		if (arg == "-d" || arg == "--data" || arg == "--data-raw") && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// parseQueryParams 解析URL查询参数
func (p *CurlRequestParser) parseQueryParams(requestURL string) (map[string]string, error) {
	params := make(map[string]string)

	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return params, err
	}

	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return params, nil
}

// IsCurlCommand 检测是否为Curl命令
func (p *CurlRequestParser) IsCurlCommand(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "curl ") || trimmed == "curl"
}
