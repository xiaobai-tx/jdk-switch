package jdk

import (
	"os"
	"path/filepath"
	"testing"
)

// 创建测试用的JDK目录结构
func setupTestJDK(t *testing.T) (string, func()) {
	// 创建临时目录作为JDK根目录
	tempDir, err := os.MkdirTemp("", "fake-jdk")
	if err != nil {
		t.Fatalf("无法创建临时JDK目录: %v", err)
	}

	// 创建bin目录
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0755); err != nil {
		t.Fatalf("无法创建bin目录: %v", err)
	}

	// 创建lib目录
	libDir := filepath.Join(tempDir, "lib")
	if err := os.Mkdir(libDir, 0755); err != nil {
		t.Fatalf("无法创建lib目录: %v", err)
	}

	// 创建java.exe和javac.exe空文件
	javaExe := filepath.Join(binDir, "java.exe")
	if err := os.WriteFile(javaExe, []byte{}, 0755); err != nil {
		t.Fatalf("无法创建java.exe: %v", err)
	}

	javacExe := filepath.Join(binDir, "javac.exe")
	if err := os.WriteFile(javacExe, []byte{}, 0755); err != nil {
		t.Fatalf("无法创建javac.exe: %v", err)
	}

	// 创建一些JAR文件
	toolsJar := filepath.Join(libDir, "tools.jar")
	if err := os.WriteFile(toolsJar, []byte{}, 0644); err != nil {
		t.Fatalf("无法创建tools.jar: %v", err)
	}

	dtJar := filepath.Join(libDir, "dt.jar")
	if err := os.WriteFile(dtJar, []byte{}, 0644); err != nil {
		t.Fatalf("无法创建dt.jar: %v", err)
	}

	// 返回清理函数
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// 测试JDK路径验证
func TestValidateJDKPath(t *testing.T) {
	// 创建测试JDK目录
	jdkPath, cleanup := setupTestJDK(t)
	defer cleanup()

	// 测试有效的JDK路径
	if !ValidateJDKPath(jdkPath) {
		t.Errorf("ValidateJDKPath(%s) 应该返回 true", jdkPath)
	}

	// 测试无效路径
	invalidPath := filepath.Join(os.TempDir(), "non-existent-jdk")
	if ValidateJDKPath(invalidPath) {
		t.Errorf("ValidateJDKPath(%s) 应该返回 false", invalidPath)
	}

	// 测试缺少java.exe的路径
	javaExePath := filepath.Join(jdkPath, "bin", "java.exe")
	os.Remove(javaExePath)
	if ValidateJDKPath(jdkPath) {
		t.Errorf("缺少java.exe时ValidateJDKPath应该返回false")
	}

	// 重新创建java.exe
	if err := os.WriteFile(javaExePath, []byte{}, 0755); err != nil {
		t.Fatalf("无法重新创建java.exe: %v", err)
	}

	// 测试缺少javac.exe的路径
	javacExePath := filepath.Join(jdkPath, "bin", "javac.exe")
	os.Remove(javacExePath)
	if ValidateJDKPath(jdkPath) {
		t.Errorf("缺少javac.exe时ValidateJDKPath应该返回false")
	}
}

// SetJavaHome 函数涉及系统环境变量修改，需要模拟或跳过
// 这里我们创建一个模拟测试版本
func TestSetJavaHomeMock(t *testing.T) {
	// 跳过在非Windows环境的真实测试
	if os.Getenv("RUN_SYSTEM_TESTS") != "true" {
		t.Skip("跳过系统环境变量测试。设置RUN_SYSTEM_TESTS=true以启用")
	}

	// 创建测试JDK目录
	jdkPath, cleanup := setupTestJDK(t)
	defer cleanup()

	// 保存原始环境变量
	origJavaHome := os.Getenv("JAVA_HOME")
	origPath := os.Getenv("PATH")
	origClasspath := os.Getenv("CLASSPATH")

	// 测试结束后恢复环境变量
	defer func() {
		os.Setenv("JAVA_HOME", origJavaHome)
		os.Setenv("PATH", origPath)
		os.Setenv("CLASSPATH", origClasspath)
	}()

	// 这里我们不真正调用SetJavaHome，因为它会修改系统环境变量
	// 而是验证JDK目录结构的正确性
	if !ValidateJDKPath(jdkPath) {
		t.Fatalf("测试JDK路径无效: %s", jdkPath)
	}

	// 检查tools.jar和dt.jar是否存在
	toolsJar := filepath.Join(jdkPath, "lib", "tools.jar")
	if _, err := os.Stat(toolsJar); os.IsNotExist(err) {
		t.Errorf("tools.jar不存在: %s", toolsJar)
	}

	dtJar := filepath.Join(jdkPath, "lib", "dt.jar")
	if _, err := os.Stat(dtJar); os.IsNotExist(err) {
		t.Errorf("dt.jar不存在: %s", dtJar)
	}
} 