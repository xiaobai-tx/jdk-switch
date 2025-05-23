package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultDir  = "C:\\jdk-switch"
	DefaultFile = "config.json"
)

type Config struct {
	JDKPaths       map[string]string `json:"jdk_paths"`
	CurrentVersion string            `json:"current_version"`
}

// InitDefaultConfig 初始化默认配置
func InitDefaultConfig() error {
	// 创建配置目录
	if err := os.MkdirAll(DefaultDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 检查配置文件是否存在
	configPath := filepath.Join(DefaultDir, DefaultFile)
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		return fmt.Errorf("配置文件已存在: %s", configPath)
	}

	// 创建默认配置
	defaultConfig := &Config{
		JDKPaths: map[string]string{
			"8":  "C:\\Program Files\\Java\\jdk1.8.0_301",
			"11": "C:\\Program Files\\Java\\jdk-11.0.12",
			"17": "C:\\Program Files\\Java\\jdk-17.0.2",
		},
		CurrentVersion: "8",
	}

	// 保存默认配置
	return defaultConfig.SaveConfig()
}

func LoadConfig() (*Config, error) {
	configPath := filepath.Join(DefaultDir, DefaultFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if len(config.JDKPaths) == 0 {
		return nil, errors.New("配置文件中没有JDK路径信息")
	}

	if config.CurrentVersion == "" {
		// 如果没有设置当前版本，使用第一个可用的版本
		for version := range config.JDKPaths {
			config.CurrentVersion = version
			break
		}
	}

	return &config, nil
}

func (c *Config) SaveConfig() error {
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	configPath := filepath.Join(DefaultDir, DefaultFile)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

func (c *Config) GetJDKPath(version string) (string, error) {
	path, exists := c.JDKPaths[version]
	if !exists {
		return "", fmt.Errorf("JDK版本 %s 不存在", version)
	}
	return path, nil
}

func (c *Config) UpdateCurrentVersion(version string) error {
	if _, exists := c.JDKPaths[version]; !exists {
		return fmt.Errorf("JDK版本 %s 不存在", version)
	}
	c.CurrentVersion = version
	return nil
}
