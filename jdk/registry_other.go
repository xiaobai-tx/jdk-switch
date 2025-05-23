//go:build !windows
// +build !windows

package jdk

import (
	"fmt"
)

// GetSystemEnvVarFromRegistry 从注册表直接读取系统环境变量原始值
// 在非Windows平台上，这个函数总是返回错误
func GetSystemEnvVarFromRegistry(name string) (string, error) {
	return "", fmt.Errorf("不支持的平台: 只有Windows支持通过注册表获取环境变量")
}

// SetSystemEnvVarToRegistry 设置系统环境变量（通过注册表）
// 在非Windows平台上，这个函数总是返回错误
func SetSystemEnvVarToRegistry(name, value string) error {
	return fmt.Errorf("不支持的平台: 只有Windows支持通过注册表设置环境变量")
}

// BroadcastEnvironmentChange 广播环境变量更改消息
// 在非Windows平台上，这个函数不执行任何操作
func BroadcastEnvironmentChange() error {
	return fmt.Errorf("不支持的平台: 只有Windows支持广播环境变量更改")
} 