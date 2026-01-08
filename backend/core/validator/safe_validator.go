package validator

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"reflect"
	"strconv"
	"strings"

	"RequestProbe/backend/core/encoding"
	"RequestProbe/backend/models"
)

// SafeValidator 安全验证器
type SafeValidator struct {
	allowedFunctions map[string]bool
	allowedOperators map[string]bool
	encodingDetector *encoding.EncodingDetector
}

// NewSafeValidator 创建安全验证器
func NewSafeValidator() *SafeValidator {
	return &SafeValidator{
		allowedFunctions: map[string]bool{
			"len":   true,
			"str":   true,
			"int":   true,
			"float": true,
			"bool":  true,
			"lower": true,
			"upper": true,
			"strip": true,
			"json":  true,
		},
		allowedOperators: map[string]bool{
			"==": true,
			"!=": true,
			"<":  true,
			"<=": true,
			">":  true,
			">=": true,
			"&&": true,
			"||": true,
			"!":  true,
			"in": true,
		},
		encodingDetector: encoding.NewEncodingDetector(),
	}
}

// ValidateExpression 验证表达式安全性
func (v *SafeValidator) ValidateExpression(expression string) error {
	if strings.TrimSpace(expression) == "" {
		return fmt.Errorf("验证表达式不能为空")
	}

	// 解析表达式为AST
	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return fmt.Errorf("表达式语法错误: %v", err)
	}

	// 检查AST节点安全性
	return v.validateASTNode(expr)
}

// validateASTNode 验证AST节点
func (v *SafeValidator) validateASTNode(node ast.Node) error {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		// 验证二元表达式
		if err := v.validateASTNode(n.X); err != nil {
			return err
		}
		if err := v.validateASTNode(n.Y); err != nil {
			return err
		}

		op := n.Op.String()
		if !v.allowedOperators[op] {
			return fmt.Errorf("不允许的操作符: %s", op)
		}

	case *ast.UnaryExpr:
		// 验证一元表达式
		if err := v.validateASTNode(n.X); err != nil {
			return err
		}

		op := n.Op.String()
		if !v.allowedOperators[op] {
			return fmt.Errorf("不允许的操作符: %s", op)
		}

	case *ast.CallExpr:
		// 验证函数调用
		if ident, ok := n.Fun.(*ast.Ident); ok {
			if !v.allowedFunctions[ident.Name] {
				return fmt.Errorf("不允许的函数: %s", ident.Name)
			}
		} else if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
			// 允许response.method()形式的调用
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "response" {
				// 验证response对象的方法调用
				if !v.isAllowedResponseMethod(sel.Sel.Name) {
					return fmt.Errorf("不允许的response方法: %s", sel.Sel.Name)
				}
			} else {
				return fmt.Errorf("不允许的方法调用")
			}
		}

		// 验证参数
		for _, arg := range n.Args {
			if err := v.validateASTNode(arg); err != nil {
				return err
			}
		}

	case *ast.SelectorExpr:
		// 验证选择器表达式 (如 response.status_code)
		if x, ok := n.X.(*ast.Ident); ok && x.Name == "response" {
			if !v.isAllowedResponseField(n.Sel.Name) {
				return fmt.Errorf("不允许的response字段: %s", n.Sel.Name)
			}
		} else {
			return fmt.Errorf("只允许访问response对象的字段")
		}

	case *ast.Ident:
		// 验证标识符
		if n.Name != "response" && !v.isBuiltinConstant(n.Name) {
			return fmt.Errorf("不允许的标识符: %s", n.Name)
		}

	case *ast.BasicLit:
		// 基本字面量是安全的
		break

	case *ast.ParenExpr:
		// 验证括号表达式
		return v.validateASTNode(n.X)

	default:
		return fmt.Errorf("不支持的表达式类型: %T", n)
	}

	return nil
}

// isAllowedResponseField 检查是否为允许的response字段
func (v *SafeValidator) isAllowedResponseField(field string) bool {
	allowedFields := map[string]bool{
		"status_code": true,
		"text":        true,
		"content":     true,
		"headers":     true,
		"cookies":     true,
		"url":         true,
		"elapsed":     true,
		"encoding":    true,
		"reason":      true,
	}
	return allowedFields[field]
}

// isAllowedResponseMethod 检查是否为允许的response方法
func (v *SafeValidator) isAllowedResponseMethod(method string) bool {
	allowedMethods := map[string]bool{
		"json": true,
	}
	return allowedMethods[method]
}

// isBuiltinConstant 检查是否为内置常量
func (v *SafeValidator) isBuiltinConstant(name string) bool {
	constants := map[string]bool{
		"true":  true,
		"false": true,
		"nil":   true,
	}
	return constants[name]
}

// EvaluateExpression 评估验证表达式（保持兼容性）
func (v *SafeValidator) EvaluateExpression(expression string, response *models.ResponseData) (bool, error) {
	// 首先验证表达式安全性
	if err := v.ValidateExpression(expression); err != nil {
		return false, err
	}

	// 创建响应对象的映射
	responseMap := v.createResponseMap(response)

	// 简单的表达式评估器
	result, err := v.evaluateSimpleExpression(expression, responseMap)
	if err != nil {
		return false, fmt.Errorf("表达式评估失败: %v", err)
	}

	return result, nil
}

// EvaluateConfig 使用新的配置系统评估响应
func (v *SafeValidator) EvaluateConfig(config *models.ValidationConfig, response *models.ResponseData) (bool, error) {
	// 如果使用自定义表达式
	if config.UseCustomExpr && config.Expression != "" {
		result, err := v.EvaluateExpression(config.Expression, response)
		return result, err
	}

	// 首先检查基本状态码：必须是2xx成功状态码
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return false, nil
	}

	// 检查文本匹配（如果启用）
	if config.TextMatching.Enabled {
		result := v.checkTextMatching(config.TextMatching, response.Body)
		return result, nil
	}

	// 检查长度范围（如果启用）
	if config.LengthRange.Enabled {
		result := v.checkLengthRange(config.LengthRange, response.Body)
		return result, nil
	}

	// 如果没有启用任何特定验证，返回详细的错误提示
	return false, fmt.Errorf("验证配置错误：未启用任何验证规则\n请在前端界面中配置以下验证方式之一：\n1. 文本匹配验证：检查响应中是否包含特定文本\n2. 长度范围验证：检查响应长度是否在指定范围内\n3. 自定义表达式验证：使用自定义表达式进行验证")
}

// checkTextMatching 检查文本匹配
func (v *SafeValidator) checkTextMatching(config models.TextMatchingConfig, responseBody string) bool {
	// 如果没有配置匹配文本，默认认为成功（只要有响应内容）
	if len(config.Texts) == 0 {
		return len(responseBody) > 0
	}

	text := responseBody
	if !config.CaseSensitive {
		text = strings.ToLower(text)
	}

	matchCount := 0
	for _, searchText := range config.Texts {
		if searchText == "" {
			continue
		}

		checkText := searchText
		if !config.CaseSensitive {
			checkText = strings.ToLower(checkText)
		}

		if strings.Contains(text, checkText) {
			matchCount++
			if config.MatchMode == "any" {
				return true // 任意匹配模式，找到一个就返回true
			}
		}
	}

	// 全部匹配模式，需要所有文本都匹配
	if config.MatchMode == "all" {
		return matchCount == len(config.Texts)
	}

	// 默认为任意匹配模式
	return matchCount > 0
}

// checkLengthRange 检查长度范围
func (v *SafeValidator) checkLengthRange(config models.LengthRangeConfig, responseBody string) bool {
	length := len(responseBody)

	if length < config.MinLength {
		return false
	}

	if config.MaxLength > 0 && length > config.MaxLength {
		return false
	}

	return true
}

// createResponseMap 创建响应数据映射
func (v *SafeValidator) createResponseMap(response *models.ResponseData) map[string]interface{} {
	responseMap := map[string]interface{}{
		"status_code": response.StatusCode,
		"text":        response.Body,
		"content":     response.Body,
		"headers":     response.Headers,
		"cookies":     response.Cookies,
		"url":         response.URL,
		"elapsed":     response.Duration,
	}

	// 添加json()方法的模拟
	if response.Body != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(response.Body), &jsonData); err == nil {
			responseMap["json"] = jsonData
		}
	}

	return responseMap
}

// evaluateSimpleExpression 简单表达式评估器
func (v *SafeValidator) evaluateSimpleExpression(expression string, responseMap map[string]interface{}) (bool, error) {
	// 这里实现一个简化的表达式评估器
	// 在实际项目中，可以使用更完善的表达式引擎

	// 替换response.字段为实际值
	expr := expression

	// 处理status_code
	if statusCode, ok := responseMap["status_code"].(int); ok {
		expr = strings.ReplaceAll(expr, "response.status_code", strconv.Itoa(statusCode))
	}

	// 处理简单的比较表达式
	if strings.Contains(expr, "==") {
		parts := strings.Split(expr, "==")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			leftVal, err := v.parseValue(left, responseMap)
			if err != nil {
				return false, err
			}

			rightVal, err := v.parseValue(right, responseMap)
			if err != nil {
				return false, err
			}

			return reflect.DeepEqual(leftVal, rightVal), nil
		}
	}

	// 处理范围比较 (如 200 <= status_code < 300)
	if strings.Contains(expr, "<=") && strings.Contains(expr, "<") {
		// 简化处理状态码范围
		if statusCode, ok := responseMap["status_code"].(int); ok {
			if statusCode >= 200 && statusCode < 300 {
				return true, nil
			}
		}
		return false, nil
	}

	// 处理包含检查
	if strings.Contains(expr, " in ") {
		parts := strings.Split(expr, " in ")
		if len(parts) == 2 {
			needle := strings.Trim(strings.TrimSpace(parts[0]), "\"'")
			haystack := strings.TrimSpace(parts[1])

			if haystack == "response.text" {
				if text, ok := responseMap["text"].(string); ok {
					return strings.Contains(text, needle), nil
				}
			}
		}
	}

	return false, fmt.Errorf("不支持的表达式格式")
}

// parseValue 解析值
func (v *SafeValidator) parseValue(value string, responseMap map[string]interface{}) (interface{}, error) {
	value = strings.TrimSpace(value)

	// 数字
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal, nil
	}

	// 字符串
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\""), nil
	}

	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return strings.Trim(value, "'"), nil
	}

	// response字段
	if strings.HasPrefix(value, "response.") {
		field := strings.TrimPrefix(value, "response.")
		if val, ok := responseMap[field]; ok {
			return val, nil
		}
	}

	return value, nil
}

// DetectEncoding 检测响应编码
func (v *SafeValidator) DetectEncoding(responseBody []byte, calibrationText string) (string, error) {
	return v.encodingDetector.DetectEncoding(responseBody, calibrationText)
}

// DecodeResponse 使用指定编码解码响应
func (v *SafeValidator) DecodeResponse(responseBody []byte, encodingName string) (string, error) {
	return v.encodingDetector.DecodeWithEncoding(responseBody, encodingName)
}

// AutoDetectEncoding 自动检测编码并转换
func (v *SafeValidator) AutoDetectEncoding(responseBody []byte) (string, string, error) {
	return v.encodingDetector.AutoDetectEncoding(responseBody)
}

// GetSupportedEncodings 获取支持的编码列表
func (v *SafeValidator) GetSupportedEncodings() []string {
	return v.encodingDetector.GetSupportedEncodings()
}

// GetCommonEncodings 获取常用编码列表
func (v *SafeValidator) GetCommonEncodings() []string {
	return v.encodingDetector.GetCommonEncodings()
}
