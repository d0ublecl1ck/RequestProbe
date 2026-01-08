package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"RequestProbe/backend/models"
)

// ExpressionManager 表达式管理器
type ExpressionManager struct {
	configDir string
	templates []models.ExpressionTemplate
}

// NewExpressionManager 创建表达式管理器
func NewExpressionManager() *ExpressionManager {
	// 获取用户配置目录
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".requestprobe")

	// 确保配置目录存在
	os.MkdirAll(configDir, 0755)

	manager := &ExpressionManager{
		configDir: configDir,
		templates: []models.ExpressionTemplate{},
	}

	// 加载默认模板
	manager.loadDefaultTemplates()

	// 加载用户自定义模板
	manager.loadUserTemplates()

	return manager
}

// loadDefaultTemplates 加载默认表达式模板（已删除所有预制模板）
func (m *ExpressionManager) loadDefaultTemplates() {
	// 不再加载任何预制模板，用户可以自定义添加
	m.templates = []models.ExpressionTemplate{}
}

// loadUserTemplates 加载用户自定义模板
func (m *ExpressionManager) loadUserTemplates() {
	templatesFile := filepath.Join(m.configDir, "templates.json")

	if _, err := os.Stat(templatesFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(templatesFile)
	if err != nil {
		return
	}

	var userTemplates []models.ExpressionTemplate
	if err := json.Unmarshal(data, &userTemplates); err != nil {
		return
	}

	m.templates = append(m.templates, userTemplates...)
}

// saveUserTemplates 保存用户自定义模板
func (m *ExpressionManager) saveUserTemplates() error {
	// 过滤出用户自定义模板（非默认模板）
	var userTemplates []models.ExpressionTemplate
	defaultIDs := map[string]bool{
		"status_success":     true,
		"status_ok":          true,
		"content_contains":   true,
		"json_status_ok":     true,
		"response_not_empty": true,
		"response_length":    true,
		"no_error_message":   true,
		"chinese_content":    true,
	}

	for _, template := range m.templates {
		if !defaultIDs[template.ID] {
			userTemplates = append(userTemplates, template)
		}
	}

	data, err := json.MarshalIndent(userTemplates, "", "  ")
	if err != nil {
		return err
	}

	templatesFile := filepath.Join(m.configDir, "templates.json")
	return os.WriteFile(templatesFile, data, 0644)
}

// GetAllTemplates 获取所有模板
func (m *ExpressionManager) GetAllTemplates() []models.ExpressionTemplate {
	return m.templates
}

// GetTemplatesByCategory 按分类获取模板
func (m *ExpressionManager) GetTemplatesByCategory(category string) []models.ExpressionTemplate {
	var result []models.ExpressionTemplate
	for _, template := range m.templates {
		if template.Category == category {
			result = append(result, template)
		}
	}
	return result
}

// GetTemplateByID 根据ID获取模板
func (m *ExpressionManager) GetTemplateByID(id string) (*models.ExpressionTemplate, error) {
	for _, template := range m.templates {
		if template.ID == id {
			return &template, nil
		}
	}
	return nil, fmt.Errorf("未找到ID为 %s 的模板", id)
}

// AddTemplate 添加新模板
func (m *ExpressionManager) AddTemplate(template models.ExpressionTemplate) error {
	// 检查ID是否已存在
	for _, existing := range m.templates {
		if existing.ID == template.ID {
			return fmt.Errorf("ID为 %s 的模板已存在", template.ID)
		}
	}

	// 生成ID（如果为空）
	if template.ID == "" {
		template.ID = fmt.Sprintf("custom_%d", time.Now().Unix())
	}

	m.templates = append(m.templates, template)
	return m.saveUserTemplates()
}

// UpdateTemplate 更新模板
func (m *ExpressionManager) UpdateTemplate(template models.ExpressionTemplate) error {
	for i, existing := range m.templates {
		if existing.ID == template.ID {
			m.templates[i] = template
			return m.saveUserTemplates()
		}
	}
	return fmt.Errorf("未找到ID为 %s 的模板", template.ID)
}

// DeleteTemplate 删除模板
func (m *ExpressionManager) DeleteTemplate(id string) error {
	// 检查是否为默认模板
	defaultIDs := map[string]bool{
		"status_success":     true,
		"status_ok":          true,
		"content_contains":   true,
		"json_status_ok":     true,
		"response_not_empty": true,
		"response_length":    true,
		"no_error_message":   true,
		"chinese_content":    true,
	}

	if defaultIDs[id] {
		return fmt.Errorf("不能删除默认模板")
	}

	for i, template := range m.templates {
		if template.ID == id {
			m.templates = append(m.templates[:i], m.templates[i+1:]...)
			return m.saveUserTemplates()
		}
	}

	return fmt.Errorf("未找到ID为 %s 的模板", id)
}

// GetCategories 获取所有分类
func (m *ExpressionManager) GetCategories() []string {
	categoryMap := make(map[string]bool)
	for _, template := range m.templates {
		categoryMap[template.Category] = true
	}

	var categories []string
	for category := range categoryMap {
		categories = append(categories, category)
	}

	return categories
}

// ExportTemplates 导出模板
func (m *ExpressionManager) ExportTemplates(filePath string) error {
	data, err := json.MarshalIndent(m.templates, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// ImportTemplates 导入模板
func (m *ExpressionManager) ImportTemplates(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var importedTemplates []models.ExpressionTemplate
	if err := json.Unmarshal(data, &importedTemplates); err != nil {
		return err
	}

	// 合并模板，避免ID冲突
	existingIDs := make(map[string]bool)
	for _, template := range m.templates {
		existingIDs[template.ID] = true
	}

	var addedCount int
	for _, template := range importedTemplates {
		if !existingIDs[template.ID] {
			m.templates = append(m.templates, template)
			addedCount++
		}
	}

	if addedCount > 0 {
		return m.saveUserTemplates()
	}

	return nil
}

// ValidateExpression 验证表达式语法
func (m *ExpressionManager) ValidateExpression(expression string) error {
	// 这里可以调用validator包的验证方法
	// 为了避免循环依赖，这里做简单验证
	if expression == "" {
		return fmt.Errorf("表达式不能为空")
	}

	// 检查是否包含response关键字
	if !strings.Contains(expression, "response") {
		return fmt.Errorf("表达式必须包含response对象")
	}

	return nil
}
