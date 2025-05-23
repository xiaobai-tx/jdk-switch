package jdk

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// BackupEnvironmentVariables 备份当前系统环境变量到C:\jdk-switch\backup\年月日时分秒目录
func BackupEnvironmentVariables() error {
	// 创建备份目录
	baseDir := `C:\jdk-switch`
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("创建备份基础目录失败: %v", err)
	}

	// 创建backup子目录
	backupBaseDir := filepath.Join(baseDir, "backup")
	if err := os.MkdirAll(backupBaseDir, 0755); err != nil {
		return fmt.Errorf("创建备份子目录失败: %v", err)
	}

	// 获取当前时间作为备份标识
	now := time.Now()
	timestamp := now.Format("20060102_150405")
	backupDir := filepath.Join(backupBaseDir, timestamp)

	// 创建时间戳子目录
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份时间目录失败: %v", err)
	}

	// 从注册表获取系统PATH环境变量（保留原始变量引用）
	pathSystem, err := GetSystemEnvVarFromRegistry("Path")
	if err != nil {
		return fmt.Errorf("获取系统PATH环境变量失败: %v", err)
	}

	// 写入PATH环境变量
	pathFile := filepath.Join(backupDir, "PATH.txt")
	if err := os.WriteFile(pathFile, []byte(pathSystem), 0644); err != nil {
		return fmt.Errorf("备份PATH环境变量失败: %v", err)
	}

	// 获取并备份系统JAVA_HOME环境变量
	javaHome, err := GetSystemEnvVarFromRegistry("JAVA_HOME")
	if err != nil {
		return fmt.Errorf("获取系统JAVA_HOME环境变量失败: %v", err)
	}

	javaHomeFile := filepath.Join(backupDir, "JAVA_HOME.txt")
	if err := os.WriteFile(javaHomeFile, []byte(javaHome), 0644); err != nil {
		return fmt.Errorf("备份JAVA_HOME环境变量失败: %v", err)
	}

	// 获取并备份系统CLASSPATH环境变量
	classpath, err := GetSystemEnvVarFromRegistry("CLASSPATH")
	if err != nil {
		return fmt.Errorf("获取系统CLASSPATH环境变量失败: %v", err)
	}

	classpathFile := filepath.Join(backupDir, "CLASSPATH.txt")
	if err := os.WriteFile(classpathFile, []byte(classpath), 0644); err != nil {
		return fmt.Errorf("备份CLASSPATH环境变量失败: %v", err)
	}

	// 创建备份信息文件
	infoContent := fmt.Sprintf("备份时间: %s\n", now.Format("2006-01-02 15:04:05"))
	infoContent += "备份文件:\n"
	infoContent += fmt.Sprintf("- PATH: %s\n", pathFile)
	infoContent += fmt.Sprintf("- JAVA_HOME: %s\n", javaHomeFile)
	infoContent += fmt.Sprintf("- CLASSPATH: %s\n", classpathFile)

	infoFile := filepath.Join(backupDir, "backup_info.txt")
	if err := os.WriteFile(infoFile, []byte(infoContent), 0644); err != nil {
		return fmt.Errorf("创建备份信息文件失败: %v", err)
	}

	// 打印备份成功信息，使用实际时间戳
	fmt.Printf("环境变量已备份到 C:\\jdk-switch\\backup\\%s 目录\n", timestamp)

	return nil
}

func SetJavaHome(jdkPath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("当前只支持Windows系统")
	}

	// 验证JDK路径是否存在
	if _, err := os.Stat(jdkPath); os.IsNotExist(err) {
		return fmt.Errorf("JDK路径不存在: %s", jdkPath)
	}

	// 注意：ValidateJDKPath已经在switchJDK函数中调用过，这里不再重复验证

	// 备份当前环境变量
	backupStart := time.Now()
	if err := BackupEnvironmentVariables(); err != nil {
		return fmt.Errorf("备份环境变量失败: %v", err)
	}
	backupDuration := time.Since(backupStart)
	fmt.Printf("备份环境变量耗时: %s\n", backupDuration)

	// 检查是否存在Oracle Java路径问题
	oracleJavaPathExists := checkOracleJavaPath()
	
	// 读取环境变量阶段开始时间
	readEnvStart := time.Now()
	
	// 获取系统级PATH环境变量
	pathSystem, err := GetSystemEnvVarFromRegistry("Path")
	if err != nil {
		return fmt.Errorf("获取系统PATH环境变量失败: %v", err)
	}
	
	// 读取环境变量阶段结束时间
	readEnvDuration := time.Since(readEnvStart)
	fmt.Printf("读取环境变量耗时: %s\n", readEnvDuration)
	
	// 修改环境变量阶段开始时间
	modifyEnvStart := time.Now()

	// 设置系统级JAVA_HOME环境变量
	if err := SetSystemEnvVarToRegistry("JAVA_HOME", jdkPath); err != nil {
		return fmt.Errorf("设置系统JAVA_HOME失败: %v", err)
	}

	// 解析PATH为条目列表
	pathEntries := strings.Split(pathSystem, ";")

	// 创建新的PATH，删除所有Java相关条目
	var newPathEntries []string
	jdkBinPath := filepath.Join(jdkPath, "bin") // 使用完整路径而不是变量引用

	for _, entry := range pathEntries {
		entry = strings.TrimSpace(entry)
		// 跳过空条目和Java相关条目，特别注意Oracle的javapath路径
		if entry == "" || 
		   entry == "%JAVA_HOME%\\bin" || 
		   strings.Contains(strings.ToLower(entry), "\\java\\") ||
		   strings.Contains(strings.ToLower(entry), "\\jdk") ||
		   strings.Contains(strings.ToLower(entry), "oracle\\java\\javapath") {
			continue
		}
		newPathEntries = append(newPathEntries, entry)
	}

	// 在PATH开头添加新的JDK bin路径（使用完整路径）
	newPathEntries = append([]string{jdkBinPath}, newPathEntries...)
	newPath := strings.Join(newPathEntries, ";")

	// 更新系统级PATH环境变量
	if err := SetSystemEnvVarToRegistry("Path", newPath); err != nil {
		return fmt.Errorf("更新系统PATH失败: %v", err)
	}

	// 设置系统级CLASSPATH环境变量
	dtJarPath := filepath.Join(jdkPath, "lib", "dt.jar")
	toolsJarPath := filepath.Join(jdkPath, "lib", "tools.jar")
	
	// 使用完整路径而不是变量引用
	classpath := fmt.Sprintf(".;%s;%s;", dtJarPath, toolsJarPath)

	// 检查JDK中是否存在这些jar文件
	var warnings []string
	if _, err := os.Stat(dtJarPath); os.IsNotExist(err) {
		warnings = append(warnings, fmt.Sprintf("警告: 文件不存在 %s", dtJarPath))
	}
	if _, err := os.Stat(toolsJarPath); os.IsNotExist(err) {
		warnings = append(warnings, fmt.Sprintf("警告: 文件不存在 %s", toolsJarPath))
	}

	// 设置CLASSPATH环境变量
	if err := SetSystemEnvVarToRegistry("CLASSPATH", classpath); err != nil {
		return fmt.Errorf("设置系统CLASSPATH失败: %v", err)
	}
	
	// 修改环境变量阶段结束时间
	modifyEnvDuration := time.Since(modifyEnvStart)
	fmt.Printf("修改环境变量耗时: %s\n", modifyEnvDuration)

	// 广播环境变量阶段开始时间
	broadcastStart := time.Now()
	
	// 所有环境变量都设置完成后，只执行一次广播
	if err := BroadcastEnvironmentChange(); err != nil {
		fmt.Printf("警告: 环境变量可能需要手动刷新 (%v)\n", err)
	} else {
		fmt.Println("\n环境变量已成功通知系统")
	}
	
	// 广播环境变量阶段结束时间
	broadcastDuration := time.Since(broadcastStart)
	fmt.Printf("广播环境变量变更耗时: %s\n", broadcastDuration)
	
	// 总耗时统计
	totalDuration := backupDuration + readEnvDuration + modifyEnvDuration + broadcastDuration
	fmt.Printf("\n总耗时: %s\n", totalDuration)

	// 如果有警告，返回警告信息但不视为错误
	if len(warnings) > 0 {
		fmt.Println(strings.Join(warnings, "\n"))
	}

	// 只保留Oracle Java路径问题的警告
	if oracleJavaPathExists {
		fmt.Println("\n警告: 检测到系统中存在Oracle Java路径(C:\\Program Files\\Common Files\\Oracle\\Java\\javapath)")
		fmt.Println("此路径可能导致java命令始终使用固定版本，而非您切换后的版本。")
		fmt.Println("建议执行以下操作：")
		fmt.Println("1. 从环境变量编辑器中手动删除此路径")
		fmt.Println("2. 或临时重命名该目录: C:\\Program Files\\Common Files\\Oracle\\Java\\javapath")
	}

	return nil
}

func ValidateJDKPath(path string) bool {
	// 检查java.exe是否存在
	javaExe := filepath.Join(path, "bin", "java.exe")
	if _, err := os.Stat(javaExe); err != nil {
		return false
	}

	// 检查javac.exe是否存在
	javacExe := filepath.Join(path, "bin", "javac.exe")
	if _, err := os.Stat(javacExe); err != nil {
		return false
	}

	return true
}

// checkOracleJavaPath 检查系统中是否存在Oracle Java路径问题
func checkOracleJavaPath() bool {
	// 检查Oracle Java路径是否存在
	oraclePath := "C:\\Program Files\\Common Files\\Oracle\\Java\\javapath"
	if _, err := os.Stat(oraclePath); err == nil {
		// 路径存在，检查其中是否有java.exe
		javaExe := filepath.Join(oraclePath, "java.exe")
		if _, err := os.Stat(javaExe); err == nil {
			// 路径中存在java.exe，可能会导致问题
			return true
		}
	}
	
	// 检查环境变量PATH中是否包含Oracle Java路径
	pathSystem, err := GetSystemEnvVarFromRegistry("Path")
	if err != nil {
		return false // 无法读取PATH，假设没有问题
	}
	
	// 检查PATH中是否包含Oracle路径
	pathEntries := strings.Split(pathSystem, ";")
	for _, entry := range pathEntries {
		entry = strings.TrimSpace(entry)
		if strings.Contains(strings.ToLower(entry), "oracle\\java\\javapath") {
			return true
		}
	}
	
	return false
}
