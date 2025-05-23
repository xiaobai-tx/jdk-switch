package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// 测试前的准备工作和测试后的清理工作
func setupTestConfig(t *testing.T) (string, *Config, func()) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "jdk-switch-test")
	if err != nil {
		t.Fatalf("无法创建临时测试目录: %v", err)
	}

	// 备份原始常量
	origConfigDir := DefaultDir
	origConfigFile := DefaultFile

	// 修改常量指向临时目录
	DefaultDir = tempDir
	DefaultFile = "config-test.json"

	// 创建测试配置
	testConfig := &Config{
		JDKPaths: map[string]string{
			"8":  "C:\\Test\\JDK8",
			"11": "C:\\Test\\JDK11",
			"17": "C:\\Test\\JDK17",
		},
		CurrentVersion: "8",
	}

	// 返回清理函数
	cleanup := func() {
		// 恢复原始常量
		DefaultDir = origConfigDir
		DefaultFile = origConfigFile

		// 删除临时目录
		os.RemoveAll(tempDir)
	}

	return tempDir, testConfig, cleanup
}

// 测试保存和加载配置
func TestSaveAndLoadConfig(t *testing.T) {
	tempDir, testConfig, cleanup := setupTestConfig(t)
	defer cleanup()

	// 测试保存配置
	if err := testConfig.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig() 错误: %v", err)
	}

	// 检查配置文件是否已创建
	configPath := filepath.Join(tempDir, DefaultFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("配置文件未创建: %s", configPath)
	}

	// 测试加载配置
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() 错误: %v", err)
	}

	// 验证加载的配置是否与保存的配置一致
	if loadedConfig.CurrentVersion != testConfig.CurrentVersion {
		t.Errorf("期望当前版本 %s, 得到 %s", testConfig.CurrentVersion, loadedConfig.CurrentVersion)
	}

	// 验证JDK路径
	for version, path := range testConfig.JDKPaths {
		loadedPath, exists := loadedConfig.JDKPaths[version]
		if !exists {
			t.Errorf("未找到JDK版本 %s", version)
			continue
		}
		if loadedPath != path {
			t.Errorf("JDK %s 路径不匹配: 期望 %s, 得到 %s", version, path, loadedPath)
		}
	}
}

// 测试初始化默认配置
func TestInitDefaultConfig(t *testing.T) {
	_, _, cleanup := setupTestConfig(t)
	defer cleanup()

	// 测试初始化配置
	if err := InitDefaultConfig(); err != nil {
		t.Fatalf("InitDefaultConfig() 错误: %v", err)
	}

	// 加载并验证初始化的配置
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("初始化后无法加载配置: %v", err)
	}

	// 验证是否包含默认JDK版本
	if len(config.JDKPaths) == 0 {
		t.Error("初始化的配置没有JDK路径")
	}

	if config.CurrentVersion == "" {
		t.Error("初始化的配置没有设置当前版本")
	}
}

// 测试获取JDK路径
func TestGetJDKPath(t *testing.T) {
	_, testConfig, cleanup := setupTestConfig(t)
	defer cleanup()

	// 测试获取有效路径
	path, err := testConfig.GetJDKPath("8")
	if err != nil {
		t.Errorf("GetJDKPath(8) 错误: %v", err)
	}
	if path != "C:\\Test\\JDK8" {
		t.Errorf("期望路径 C:\\Test\\JDK8, 得到 %s", path)
	}

	// 测试获取无效版本
	_, err = testConfig.GetJDKPath("999")
	if err == nil {
		t.Error("获取无效版本应该返回错误")
	}
}

// 测试更新当前版本
func TestUpdateCurrentVersion(t *testing.T) {
	_, testConfig, cleanup := setupTestConfig(t)
	defer cleanup()

	// 测试更新到有效版本
	if err := testConfig.UpdateCurrentVersion("11"); err != nil {
		t.Errorf("UpdateCurrentVersion(11) 错误: %v", err)
	}
	if testConfig.CurrentVersion != "11" {
		t.Errorf("更新当前版本失败: 期望 11, 得到 %s", testConfig.CurrentVersion)
	}

	// 测试更新到无效版本
	if err := testConfig.UpdateCurrentVersion("999"); err == nil {
		t.Error("更新到无效版本应该返回错误")
	}
} 