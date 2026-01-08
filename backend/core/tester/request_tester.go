package tester

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"RequestProbe/backend/core/validator"
	"RequestProbe/backend/models"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

// RequestTester 请求测试器
type RequestTester struct {
	client    *http.Client
	Validator *validator.SafeValidator // 导出字段
}

// NewRequestTester 创建请求测试器
func NewRequestTester() *RequestTester {
	return &RequestTester{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Validator: validator.NewSafeValidator(),
	}
}

// SetTimeout 设置请求超时时间
func (t *RequestTester) SetTimeout(timeout time.Duration) {
	t.client.Timeout = timeout
}

// SetProxy 设置代理
func (t *RequestTester) SetProxy(proxyURL string) error {
	if proxyURL == "" {
		t.client.Transport = nil
		return nil
	}

	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("无效的代理URL: %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	t.client.Transport = transport
	return nil
}

// TestRequest 测试单个请求
func (t *RequestTester) TestRequest(req *models.ParsedRequest, config *models.ValidationConfig) (*models.ResponseData, error) {
	// 创建HTTP请求
	httpReq, err := t.createHTTPRequest(req)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 执行请求
	start := time.Now()
	resp, err := t.client.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("请求执行失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 自动检测并转换编码
	decodedBody, detectedEncoding := t.autoDetectAndDecodeResponse(body, resp.Header.Get("Content-Type"))
	if decodedBody == "" {
		decodedBody = string(body) // 如果自动检测失败，使用原始内容
	}

	// 构建响应数据
	responseData := &models.ResponseData{
		StatusCode:       resp.StatusCode,
		Headers:          make(map[string]string),
		Body:             decodedBody, // 使用解码后的内容
		Cookies:          resp.Cookies(),
		URL:              resp.Request.URL.String(),
		Duration:         duration,
		ContentLength:    int64(len(body)),         // 原始字节长度
		CharacterCount:   len([]rune(decodedBody)), // 解码后字符长度
		RawBody:          body,                     // 保存原始字节数据
		DetectedEncoding: detectedEncoding,         // 保存检测到的编码
	}

	// 转换响应头
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseData.Headers[key] = values[0]
		}
	}

	return responseData, nil
}

// ValidateResponse 验证响应（保持兼容性）
func (t *RequestTester) ValidateResponse(response *models.ResponseData, expression string) (bool, error) {
	if expression == "" {
		// 默认验证：状态码在200-299范围内
		return response.StatusCode >= 200 && response.StatusCode < 300, nil
	}

	return t.Validator.EvaluateExpression(expression, response)
}

// ValidateResponseWithConfig 使用新配置验证响应
func (t *RequestTester) ValidateResponseWithConfig(response *models.ResponseData, config *models.ValidationConfig) (bool, error) {
	return t.Validator.EvaluateConfig(config, response)
}

// TestFieldNecessity 测试字段必要性（带重试机制）
func (t *RequestTester) TestFieldNecessity(originalReq *models.ParsedRequest, fieldName, fieldType string, config *models.ValidationConfig) (*models.TestResult, error) {
	// 创建测试请求（移除指定字段）
	testReq := t.createTestRequest(originalReq, fieldName, fieldType)

	// 执行测试请求（带重试）
	response, err := t.TestRequestWithRetry(testReq, config)
	if err != nil {
		return &models.TestResult{
			FieldName:  fieldName,
			FieldType:  fieldType,
			IsRequired: true,
			TestPassed: false,
			ErrorMsg:   err.Error(),
		}, nil
	}

	// 使用新的验证配置
	passed, err := t.ValidateResponseWithConfig(response, config)
	if err != nil {
		return &models.TestResult{
			FieldName:   fieldName,
			FieldType:   fieldType,
			IsRequired:  true,
			TestPassed:  false,
			ErrorMsg:    fmt.Sprintf("验证失败: %v", err),
			StatusCode:  response.StatusCode,
			ResponseMsg: response.Body,
		}, nil
	}

	return &models.TestResult{
		FieldName:   fieldName,
		FieldType:   fieldType,
		IsRequired:  !passed, // 如果移除字段后测试失败，说明字段是必需的
		TestPassed:  passed,
		StatusCode:  response.StatusCode,
		ResponseMsg: response.Body,
	}, nil
}

// TestRequestWithRetry 带重试机制的请求测试
func (t *RequestTester) TestRequestWithRetry(req *models.ParsedRequest, config *models.ValidationConfig) (*models.ResponseData, error) {
	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3 // 默认重试3次
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err := t.TestRequest(req, config)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// 如果不是最后一次尝试，等待一段时间再重试
		if attempt < maxRetries {
			// 指数退避：100ms, 200ms, 400ms...
			waitTime := time.Duration(100*(1<<attempt)) * time.Millisecond
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("重试 %d 次后仍然失败: %v", maxRetries, lastErr)
}

// createHTTPRequest 创建HTTP请求
func (t *RequestTester) createHTTPRequest(req *models.ParsedRequest) (*http.Request, error) {
	var body io.Reader
	if req.Body != "" {
		body = bytes.NewBufferString(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return nil, err
	}

	// 清空默认头部，确保只使用我们明确设置的头部
	httpReq.Header = make(http.Header)

	// 设置Headers
	for key, value := range req.Headers {
		// 特殊处理User-Agent：如果值为空字符串，则完全不设置这个header
		if strings.ToLower(key) == "user-agent" && value == "" {
			continue // 跳过，不设置User-Agent header
		}
		httpReq.Header.Set(key, value)
	}

	// 设置Cookies
	for name, value := range req.Cookies {
		cookie := &http.Cookie{
			Name:  name,
			Value: value,
		}
		httpReq.AddCookie(cookie)
	}

	return httpReq, nil
}

// createTestRequest 创建测试请求（移除指定字段）
func (t *RequestTester) createTestRequest(original *models.ParsedRequest, fieldName, fieldType string) *models.ParsedRequest {
	// 深拷贝原始请求
	testReq := &models.ParsedRequest{
		Method:      original.Method,
		URL:         original.URL,
		Headers:     make(map[string]string),
		Cookies:     make(map[string]string),
		Body:        original.Body,
		QueryParams: make(map[string]string),
		ContentType: original.ContentType,
	}

	// 拷贝Headers（除了要测试的字段）
	for key, value := range original.Headers {
		if fieldType == "header" && key == fieldName {
			continue // 跳过要测试的header字段
		}
		testReq.Headers[key] = value
	}

	// 拷贝Cookies（除了要测试的字段）
	for key, value := range original.Cookies {
		if fieldType == "cookie" && key == fieldName {
			continue // 跳过要测试的cookie字段
		}
		testReq.Cookies[key] = value
	}

	// 拷贝查询参数
	for key, value := range original.QueryParams {
		testReq.QueryParams[key] = value
	}

	return testReq
}

// BatchTestFieldNecessity 批量测试字段必要性（累积移除算法）
func (t *RequestTester) BatchTestFieldNecessity(req *models.ParsedRequest, config *models.ValidationConfig, progressCallback func(*models.TestProgress)) (*models.BatchTestResult, error) {
	start := time.Now()

	result := &models.BatchTestResult{
		OriginalRequest: req,
		HeaderResults:   []models.TestResult{},
		CookieResults:   []models.TestResult{},
	}

	// 计算总测试数
	totalTests := len(req.Headers) + len(req.Cookies) + 1 // +1 for original request test
	result.TotalTests = totalTests
	currentStep := 0

	// 更新进度
	updateProgress := func(message string) {
		if progressCallback != nil {
			progress := &models.TestProgress{
				CurrentStep:    message,
				TotalSteps:     totalTests,
				CompletedSteps: currentStep,
				Progress:       float64(currentStep) / float64(totalTests) * 100,
				Message:        message,
			}
			progressCallback(progress)
		}
	}

	// 更新进度并发送字段测试结果
	updateProgressWithResult := func(message string, fieldResult *models.TestResult) {
		if progressCallback != nil {
			progress := &models.TestProgress{
				CurrentStep:    message,
				TotalSteps:     totalTests,
				CompletedSteps: currentStep,
				Progress:       float64(currentStep) / float64(totalTests) * 100,
				Message:        message,
				FieldResult:    fieldResult,
			}
			progressCallback(progress)
		}
	}

	// 首先测试原始请求
	updateProgress("测试原始请求...")
	originalResponse, err := t.TestRequestWithRetry(req, config)
	if err != nil {
		result.OriginalPassed = false
		result.OriginalError = err.Error()
		// 提供更详细的错误信息
		detailedError := fmt.Sprintf("原始请求测试失败: %v\n请检查:\n1. 网络连接是否正常\n2. 请求URL是否正确\n3. 代理设置是否正确\n4. 超时设置是否合理", err)
		return result, fmt.Errorf(detailedError)
	}

	// 使用新的验证配置验证原始请求
	passed, err := t.ValidateResponseWithConfig(originalResponse, config)
	if err != nil {
		result.OriginalPassed = false
		result.OriginalError = fmt.Sprintf("原始请求验证失败: %v", err)
		return result, err
	}

	result.OriginalPassed = passed
	if !passed {
		result.OriginalError = "原始请求未通过验证条件"
		return result, fmt.Errorf("原始请求未通过验证，无法继续测试")
	}

	currentStep++

	// 使用累积移除算法测试字段
	cumulativeResults, legacyResults := t.testFieldsWithCumulativeRemoval(req, config, updateProgress, updateProgressWithResult, &currentStep)

	// 设置累积测试结果
	result.CumulativeResults = cumulativeResults

	// 转换为传统格式以保持兼容性
	result.HeaderResults = legacyResults.HeaderResults
	result.CookieResults = legacyResults.CookieResults
	result.PassedTests = legacyResults.PassedTests

	// 生成简化请求
	result.SimplifiedRequest = t.generateSimplifiedRequestFromCumulative(req, cumulativeResults)
	result.SimplifiedCode = t.generateSimplifiedPythonCode(result.SimplifiedRequest)
	result.TestDuration = time.Since(start)

	updateProgress("测试完成")
	return result, nil
}

// generateSimplifiedRequest 生成简化请求
func (t *RequestTester) generateSimplifiedRequest(original *models.ParsedRequest, result *models.BatchTestResult) *models.ParsedRequest {
	simplified := &models.ParsedRequest{
		Method:      original.Method,
		URL:         original.URL,
		Headers:     make(map[string]string),
		Cookies:     make(map[string]string),
		Body:        original.Body,
		QueryParams: make(map[string]string),
		ContentType: original.ContentType,
	}

	// 只保留必需的Headers
	for _, headerResult := range result.HeaderResults {
		if headerResult.IsRequired {
			if value, exists := original.Headers[headerResult.FieldName]; exists {
				simplified.Headers[headerResult.FieldName] = value
			}
		}
	}

	// 只保留必需的Cookies
	for _, cookieResult := range result.CookieResults {
		if cookieResult.IsRequired {
			if value, exists := original.Cookies[cookieResult.FieldName]; exists {
				simplified.Cookies[cookieResult.FieldName] = value
			}
		}
	}

	// 保留所有查询参数
	for key, value := range original.QueryParams {
		simplified.QueryParams[key] = value
	}

	return simplified
}

// generateSimplifiedPythonCode 生成简化的Python代码
func (t *RequestTester) generateSimplifiedPythonCode(req *models.ParsedRequest) string {
	var code strings.Builder

	code.WriteString("import requests\n\n")

	// Headers (只包含必需的)
	if len(req.Headers) > 0 {
		code.WriteString("headers = {\n")
		for key, value := range req.Headers {
			if strings.ToLower(key) != "cookie" {
				code.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", key, value))
			}
		}
		code.WriteString("}\n")
	}

	// Cookies (只包含必需的)
	if len(req.Cookies) > 0 {
		code.WriteString("cookies = {\n")
		for key, value := range req.Cookies {
			code.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", key, value))
		}
		code.WriteString("}\n")
	}

	// 解析URL和参数
	baseURL, queryParams := t.parseURLAndParams(req.URL)
	code.WriteString(fmt.Sprintf("url = \"%s\"\n", baseURL))

	// 查询参数 (只包含必需的)
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
func (t *RequestTester) parseURLAndParams(fullURL string) (string, map[string]string) {
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

// testFieldsConcurrently 并发测试字段
func (t *RequestTester) testFieldsConcurrently(req *models.ParsedRequest, fields map[string]string, fieldType string, config *models.ValidationConfig, updateProgress func(string), currentStep *int) []models.TestResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]models.TestResult, 0, len(fields))

	// 限制并发数量，避免过多请求
	semaphore := make(chan struct{}, 5) // 最多5个并发

	for fieldName := range fields {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 更新进度
			updateProgress(fmt.Sprintf("测试%s: %s", fieldType, name))

			testResult, err := t.TestFieldNecessity(req, name, fieldType, config)
			if err != nil {
				testResult = &models.TestResult{
					FieldName:  name,
					FieldType:  fieldType,
					IsRequired: true,
					TestPassed: false,
					ErrorMsg:   err.Error(),
				}
			}

			// 线程安全地添加结果
			mu.Lock()
			results = append(results, *testResult)
			*currentStep++
			mu.Unlock()
		}(fieldName)
	}

	wg.Wait()
	return results
}

// testFieldsWithCumulativeRemoval 使用累积移除算法测试字段
func (t *RequestTester) testFieldsWithCumulativeRemoval(originalReq *models.ParsedRequest, config *models.ValidationConfig, updateProgress func(string), updateProgressWithResult func(string, *models.TestResult), currentStep *int) (*models.TestResults, *struct {
	HeaderResults []models.TestResult
	CookieResults []models.TestResult
	PassedTests   int
}) {
	// 创建累积测试状态
	cumulativeState := &models.CumulativeTestState{
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	// 深拷贝原始请求数据
	for k, v := range originalReq.Headers {
		cumulativeState.Headers[k] = v
	}
	for k, v := range originalReq.Cookies {
		cumulativeState.Cookies[k] = v
	}

	// 创建结果结构
	cumulativeResults := &models.TestResults{
		Headers: make(map[string]*models.FieldTestResult),
		Cookies: make(map[string]*models.FieldTestResult),
	}

	// 用于兼容性的传统结果
	legacyResults := &struct {
		HeaderResults []models.TestResult
		CookieResults []models.TestResult
		PassedTests   int
	}{
		HeaderResults: []models.TestResult{},
		CookieResults: []models.TestResult{},
		PassedTests:   0,
	}

	// 按原始顺序测试Headers（累积移除算法）
	headerOrder := t.getOriginalHeaderOrder(originalReq)
	for _, headerName := range headerOrder {
		// 更新进度
		updateProgress(fmt.Sprintf("测试Header: %s", headerName))

		// 检查字段是否还存在于累积状态中
		removedValue, exists := cumulativeState.Headers[headerName]
		if !exists {
			// 字段已在之前的测试中被移除，跳过
			continue
		}

		// 检查是否为User-Agent且配置为保留
		isUserAgent := strings.ToLower(headerName) == "user-agent"
		if isUserAgent && config.PreserveUserAgent {
			// 如果配置为保留User-Agent，跳过测试，直接标记为必需
			cumulativeResults.Headers[headerName] = &models.FieldTestResult{
				Required:   true, // 强制标记为必需
				Value:      removedValue,
				TestResult: &models.SingleRequestResult{Success: true}, // 假设成功
			}

			// 记录传统测试结果
			legacyResult := models.TestResult{
				FieldName:  headerName,
				FieldType:  "header",
				IsRequired: true, // 强制标记为必需
				TestPassed: true,
				ErrorMsg:   "",
			}
			legacyResults.HeaderResults = append(legacyResults.HeaderResults, legacyResult)
			legacyResults.PassedTests++

			*currentStep++

			// 立即发送包含字段测试结果的进度更新
			updateProgressWithResult(fmt.Sprintf("完成Header: %s (保留)", headerName), &legacyResult)
			continue
		}

		// 特殊处理User-Agent：设置为空字符串而不是删除
		if isUserAgent {
			cumulativeState.Headers[headerName] = ""
		} else {
			// 临时从累积状态中移除当前字段
			delete(cumulativeState.Headers, headerName)
		}

		// 构建测试请求（基于当前累积状态）
		testRequest := t.buildRequestFromState(cumulativeState, originalReq)

		// 执行测试
		testResult := t.executeRequest(testRequest, config)

		// 判断字段是否必需
		isRequired := !testResult.Success

		if isRequired {
			// 字段是必需的，恢复到累积状态中
			cumulativeState.Headers[headerName] = removedValue
		} else if !isUserAgent {
			// 如果字段不是必需的且不是User-Agent，则保持从累积状态中移除
			// User-Agent已经设置为空字符串，保持这个状态
		}

		// 记录累积测试结果
		cumulativeResults.Headers[headerName] = &models.FieldTestResult{
			Required:   isRequired,
			Value:      removedValue,
			TestResult: testResult,
		}

		// 记录传统测试结果
		legacyResult := models.TestResult{
			FieldName:  headerName,
			FieldType:  "header",
			IsRequired: isRequired,
			TestPassed: testResult.Success,
			ErrorMsg:   testResult.Error,
		}
		if testResult.ResponseInfo != nil {
			legacyResult.StatusCode = testResult.ResponseInfo.StatusCode
		}
		legacyResults.HeaderResults = append(legacyResults.HeaderResults, legacyResult)

		if testResult.Success {
			legacyResults.PassedTests++
		}

		*currentStep++

		// 立即发送包含字段测试结果的进度更新
		updateProgressWithResult(fmt.Sprintf("完成Header: %s", headerName), &legacyResult)
	}

	// 按原始顺序测试Cookies（累积移除算法）
	cookieOrder := t.getOriginalCookieOrder(originalReq)
	for _, cookieName := range cookieOrder {
		updateProgress(fmt.Sprintf("测试Cookie: %s", cookieName))

		// 检查字段是否还存在于累积状态中
		removedValue, exists := cumulativeState.Cookies[cookieName]
		if !exists {
			// 字段已在之前的测试中被移除，跳过
			continue
		}

		// 临时从累积状态中移除当前字段
		delete(cumulativeState.Cookies, cookieName)

		// 构建测试请求（基于当前累积状态）
		testRequest := t.buildRequestFromState(cumulativeState, originalReq)
		testResult := t.executeRequest(testRequest, config)

		// 判断字段是否必需
		isRequired := !testResult.Success

		if isRequired {
			// 字段是必需的，恢复到累积状态中
			cumulativeState.Cookies[cookieName] = removedValue
		}
		// 如果字段不是必需的，则保持从累积状态中移除

		// 记录累积测试结果
		cumulativeResults.Cookies[cookieName] = &models.FieldTestResult{
			Required:   isRequired,
			Value:      removedValue,
			TestResult: testResult,
		}

		// 记录传统测试结果
		legacyResult := models.TestResult{
			FieldName:  cookieName,
			FieldType:  "cookie",
			IsRequired: isRequired,
			TestPassed: testResult.Success,
			ErrorMsg:   testResult.Error,
		}
		if testResult.ResponseInfo != nil {
			legacyResult.StatusCode = testResult.ResponseInfo.StatusCode
		}
		legacyResults.CookieResults = append(legacyResults.CookieResults, legacyResult)

		if testResult.Success {
			legacyResults.PassedTests++
		}

		*currentStep++

		// 立即发送包含字段测试结果的进度更新
		updateProgressWithResult(fmt.Sprintf("完成Cookie: %s", cookieName), &legacyResult)
	}

	return cumulativeResults, legacyResults
}

// getOriginalHeaderOrder 获取原始Header顺序
func (t *RequestTester) getOriginalHeaderOrder(req *models.ParsedRequest) []string {
	order := make([]string, 0, len(req.Headers))
	for headerName := range req.Headers {
		order = append(order, headerName)
	}
	return order
}

// getOriginalCookieOrder 获取原始Cookie顺序
func (t *RequestTester) getOriginalCookieOrder(req *models.ParsedRequest) []string {
	order := make([]string, 0, len(req.Cookies))
	for cookieName := range req.Cookies {
		order = append(order, cookieName)
	}
	return order
}

// buildRequestFromState 从累积状态构建请求
func (t *RequestTester) buildRequestFromState(state *models.CumulativeTestState, original *models.ParsedRequest) *models.ParsedRequest {
	testRequest := &models.ParsedRequest{
		Method:      original.Method,
		URL:         original.URL,
		Headers:     make(map[string]string),
		Cookies:     make(map[string]string),
		Body:        original.Body,
		QueryParams: make(map[string]string),
		ContentType: original.ContentType,
	}

	// 复制累积状态中的headers
	for k, v := range state.Headers {
		testRequest.Headers[k] = v
	}

	// 复制累积状态中的cookies
	for k, v := range state.Cookies {
		testRequest.Cookies[k] = v
	}

	// 复制查询参数（保持不变）
	for k, v := range original.QueryParams {
		testRequest.QueryParams[k] = v
	}

	return testRequest
}

// 全局测试计数器
var testCounter int

// executeRequest 执行请求并返回结果
func (t *RequestTester) executeRequest(request *models.ParsedRequest, config *models.ValidationConfig) *models.SingleRequestResult {
	// 增加测试计数器
	testCounter++

	// 按照用户要求的格式打印日志
	fmt.Printf("\n=========== 第%d次测试 ===========\n", testCounter)

	// 打印headers
	fmt.Printf("headers：{")
	headerCount := 0
	for name, value := range request.Headers {
		if headerCount > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("\"%s\": \"%s\"", name, value)
		headerCount++
	}
	fmt.Printf("}\n")

	// 打印cookies
	fmt.Printf("cookies：{")
	cookieCount := 0
	for name, value := range request.Cookies {
		if cookieCount > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("\"%s\": \"%s\"", name, value)
		cookieCount++
	}
	fmt.Printf("}\n")

	// 创建HTTP请求以检查实际发送的headers
	httpReq, err := t.createHTTPRequest(request)
	if err != nil {
		fmt.Printf("表达式求值：创建请求失败 - %s\n", err.Error())
		fmt.Printf("返回包前100字符：无\n")
		return &models.SingleRequestResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// 打印实际发送的headers（包括Go自动添加的默认headers）
	fmt.Printf("实际发送的headers：{")
	actualHeaderCount := 0
	for name, values := range httpReq.Header {
		if actualHeaderCount > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("\"%s\": \"%s\"", name, values[0])
		actualHeaderCount++
	}
	fmt.Printf("}\n")

	// 额外检查：打印Go可能自动添加的headers
	fmt.Printf("Go可能自动添加的headers：\n")
	fmt.Printf("  Host: %s\n", httpReq.Host)
	fmt.Printf("  URL: %s\n", httpReq.URL.String())
	fmt.Printf("  Method: %s\n", httpReq.Method)

	// 发送HTTP请求
	response, err := t.TestRequestWithRetry(request, config)
	if err != nil {
		fmt.Printf("表达式求值：请求失败 - %s\n", err.Error())
		fmt.Printf("返回包前100字符：无\n")
		return &models.SingleRequestResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// 执行验证
	validationResult, err := t.ValidateResponseWithConfig(response, config)
	if err != nil {
		fmt.Printf("表达式求值：验证失败 - %s\n", err.Error())
		fmt.Printf("返回包前100字符：%s\n", truncateString(response.Body, 100))
		return &models.SingleRequestResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// 打印表达式求值结果
	fmt.Printf("表达式求值：%t\n", validationResult)

	// 打印返回包前100字符
	fmt.Printf("返回包前100字符：%s\n", truncateString(response.Body, 100))

	// 构建响应信息
	responseInfo := &models.ResponseInfo{
		StatusCode: response.StatusCode,
		URL:        response.URL,
		Headers:    response.Headers,
	}

	return &models.SingleRequestResult{
		Success:      validationResult,
		ResponseInfo: responseInfo,
	}
}

// generateSimplifiedRequestFromCumulative 从累积结果生成简化请求
func (t *RequestTester) generateSimplifiedRequestFromCumulative(original *models.ParsedRequest, results *models.TestResults) *models.ParsedRequest {
	simplified := &models.ParsedRequest{
		Method:      original.Method,
		URL:         original.URL,
		Headers:     make(map[string]string),
		Cookies:     make(map[string]string),
		Body:        original.Body,
		QueryParams: make(map[string]string),
		ContentType: original.ContentType,
	}

	// 只保留必需的Headers
	for headerName, result := range results.Headers {
		if result.Required {
			if value, exists := original.Headers[headerName]; exists {
				simplified.Headers[headerName] = value
			}
		}
	}

	// 只保留必需的Cookies
	for cookieName, result := range results.Cookies {
		if result.Required {
			if value, exists := original.Cookies[cookieName]; exists {
				simplified.Cookies[cookieName] = value
			}
		}
	}

	// 保留所有查询参数
	for key, value := range original.QueryParams {
		simplified.QueryParams[key] = value
	}

	return simplified
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// autoDetectAndDecodeResponse 自动检测编码并解码响应
func (t *RequestTester) autoDetectAndDecodeResponse(body []byte, contentType string) (string, string) {
	// 使用charset包自动检测编码
	encoding, name, certain := charset.DetermineEncoding(body, contentType)

	fmt.Printf("自动检测编码: %s (确定性: %v, Content-Type: %s)\n", name, certain, contentType)

	// 如果检测到的编码不是UTF-8，进行转换
	if name != "utf-8" && name != "" {
		decoder := encoding.NewDecoder()
		reader := transform.NewReader(bytes.NewReader(body), decoder)

		decoded, err := io.ReadAll(reader)
		if err != nil {
			fmt.Printf("编码转换失败: %v\n", err)
			return string(body), name // 返回原始内容和检测到的编码名
		}

		return string(decoded), name
	}

	// 如果是UTF-8或检测失败，直接返回原始内容
	return string(body), name
}
