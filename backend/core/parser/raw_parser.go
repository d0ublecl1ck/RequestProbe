package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"RequestProbe/backend/models"
)

// RawRequestParser Raw HTTP请求解析器
type RawRequestParser struct{}

// NewRawRequestParser 创建Raw请求解析器
func NewRawRequestParser() *RawRequestParser {
	return &RawRequestParser{}
}

// Parse 解析Raw HTTP请求
func (p *RawRequestParser) Parse(rawRequest string) (*models.ParsedRequest, error) {
	if strings.TrimSpace(rawRequest) == "" {
		return nil, fmt.Errorf("请求内容不能为空")
	}

	lines := strings.Split(strings.ReplaceAll(rawRequest, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("无效的请求格式")
	}

	// 解析请求行
	requestLine := strings.TrimSpace(lines[0])
	method, requestURL, err := p.parseRequestLine(requestLine)
	if err != nil {
		return nil, fmt.Errorf("解析请求行失败: %v", err)
	}

	// 解析Headers和Body
	headers := make(map[string]string)
	cookies := make(map[string]string)
	var body string
	var bodyStartIndex int

	// 查找空行，分离headers和body
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			bodyStartIndex = i + 1
			break
		}

		// 解析header
		if colonIndex := strings.Index(line, ":"); colonIndex > 0 {
			key := strings.TrimSpace(line[:colonIndex])
			value := strings.TrimSpace(line[colonIndex+1:])
			headers[key] = value

			// 特殊处理Cookie header
			if strings.ToLower(key) == "cookie" {
				cookieMap := p.parseCookieHeader(value)
				for k, v := range cookieMap {
					cookies[k] = v
				}
			}
		}
	}

	// 解析Body
	if bodyStartIndex < len(lines) {
		bodyLines := lines[bodyStartIndex:]
		body = strings.Join(bodyLines, "\n")
		body = strings.TrimSpace(body)
	}

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

// parseRequestLine 解析请求行
func (p *RawRequestParser) parseRequestLine(line string) (method, url string, err error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("无效的请求行格式")
	}

	method = strings.ToUpper(parts[0])
	url = parts[1]

	// 验证HTTP方法
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	isValidMethod := false
	for _, validMethod := range validMethods {
		if method == validMethod {
			isValidMethod = true
			break
		}
	}

	if !isValidMethod {
		return "", "", fmt.Errorf("不支持的HTTP方法: %s", method)
	}

	return method, url, nil
}

// parseCookieHeader 解析Cookie header
func (p *RawRequestParser) parseCookieHeader(cookieHeader string) map[string]string {
	cookies := make(map[string]string)

	// Cookie格式: name1=value1; name2=value2
	pairs := strings.Split(cookieHeader, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if equalIndex := strings.Index(pair, "="); equalIndex > 0 {
			name := strings.TrimSpace(pair[:equalIndex])
			value := strings.TrimSpace(pair[equalIndex+1:])
			cookies[name] = value
		}
	}

	return cookies
}

// parseQueryParams 解析URL查询参数
func (p *RawRequestParser) parseQueryParams(requestURL string) (map[string]string, error) {
	params := make(map[string]string)

	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return params, err
	}

	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			params[key] = values[0] // 取第一个值
		}
	}

	return params, nil
}

// IsRawRequest 检测是否为Raw HTTP请求格式
func (p *RawRequestParser) IsRawRequest(input string) bool {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	if len(lines) == 0 {
		return false
	}

	// 检查第一行是否为HTTP请求行格式
	firstLine := strings.TrimSpace(lines[0])
	httpMethodPattern := `^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+\S+(\s+HTTP/\d\.\d)?$`
	matched, _ := regexp.MatchString(httpMethodPattern, firstLine)

	return matched
}
