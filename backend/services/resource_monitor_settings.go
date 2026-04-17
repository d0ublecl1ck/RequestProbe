package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"RequestProbe/backend/models"
)

type resourceMonitorSettingsFile struct {
	ResourceMonitorSaveRoot string `json:"resourceMonitorSaveRoot,omitempty"`
}

func (s *ResourceMonitorService) GetSettings(ctx context.Context) (*models.ResourceMonitorSettings, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadSettings()
}

func (s *ResourceMonitorService) UpdateSaveRoot(ctx context.Context, saveRoot string) (*models.ResourceMonitorSettings, error) {
	_ = ctx
	normalizedRoot, err := normalizeSaveRootDir(saveRoot)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.readSettingsFile()
	if err != nil {
		return nil, err
	}
	settings.ResourceMonitorSaveRoot = normalizedRoot
	if err := s.writeSettingsFile(settings); err != nil {
		return nil, err
	}

	return s.loadSettings()
}

func (s *ResourceMonitorService) ResetSaveRoot(ctx context.Context) (*models.ResourceMonitorSettings, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.readSettingsFile()
	if err != nil {
		return nil, err
	}
	settings.ResourceMonitorSaveRoot = ""
	if err := s.writeSettingsFile(settings); err != nil {
		return nil, err
	}

	return s.loadSettings()
}

func (s *ResourceMonitorService) getEffectiveSaveRootDir() (string, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return "", err
	}
	return settings.SaveRootDir, nil
}

func (s *ResourceMonitorService) loadSettings() (*models.ResourceMonitorSettings, error) {
	settingsFile, err := s.readSettingsFile()
	if err != nil {
		return nil, err
	}

	defaultRoot, err := s.defaultRootFn()
	if err != nil {
		return nil, fmt.Errorf("获取默认保存目录失败: %w", err)
	}

	effectiveRoot := settingsFile.ResourceMonitorSaveRoot
	if strings.TrimSpace(effectiveRoot) == "" {
		effectiveRoot = defaultRoot
	}

	normalizedRoot, err := normalizeSaveRootDir(effectiveRoot)
	if err != nil {
		return nil, err
	}

	normalizedDefault, err := normalizeSaveRootDir(defaultRoot)
	if err != nil {
		return nil, err
	}

	return &models.ResourceMonitorSettings{
		SaveRootDir:        normalizedRoot,
		DefaultSaveRootDir: normalizedDefault,
	}, nil
}

func (s *ResourceMonitorService) readSettingsFile() (*resourceMonitorSettingsFile, error) {
	settingsPath, err := s.settingsPathFn()
	if err != nil {
		return nil, fmt.Errorf("获取资源监听设置文件失败: %w", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &resourceMonitorSettingsFile{}, nil
		}
		return nil, fmt.Errorf("读取资源监听设置失败: %w", err)
	}

	var settings resourceMonitorSettingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("解析资源监听设置失败: %w", err)
	}
	return &settings, nil
}

func (s *ResourceMonitorService) writeSettingsFile(settings *resourceMonitorSettingsFile) error {
	settingsPath, err := s.settingsPathFn()
	if err != nil {
		return fmt.Errorf("获取资源监听设置文件失败: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("创建设置目录失败: %w", err)
	}

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化资源监听设置失败: %w", err)
	}

	if err := os.WriteFile(settingsPath, payload, 0o644); err != nil {
		return fmt.Errorf("写入资源监听设置失败: %w", err)
	}
	return nil
}

func defaultResourceMonitorSettingsPath() (string, error) {
	configDir, err := appConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "settings.json"), nil
}

func defaultResourceMonitorSaveRootDir() (string, error) {
	configDir, err := appConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "downloads"), nil
}

func appConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取应用配置目录失败: %w", err)
	}
	return filepath.Join(configDir, "RequestProbe"), nil
}

func normalizeSaveRootDir(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("保存目录不能为空")
	}

	absPath, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("解析保存目录失败: %w", err)
	}

	if err := os.MkdirAll(absPath, 0o755); err != nil {
		return "", fmt.Errorf("创建保存目录失败: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("读取保存目录失败: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("保存路径必须是目录")
	}

	return filepath.Clean(absPath), nil
}
