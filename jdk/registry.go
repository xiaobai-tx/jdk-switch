//go:build windows
// +build windows

package jdk

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os/exec"
	"strings"
)

// 系统环境变量注册表路径
const envRegistryPath = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`

// GetSystemEnvVarFromRegistry 从注册表直接读取系统环境变量原始值
func GetSystemEnvVarFromRegistry(name string) (string, error) {
	// 打开系统环境变量注册表键
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, envRegistryPath, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("打开注册表失败: %v", err)
	}
	defer key.Close()

	// 读取环境变量值
	value, _, err := key.GetStringValue(name)
	if err != nil {
		// 如果键不存在，返回空字符串但不报错
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("读取环境变量值失败: %v", err)
	}

	return value, nil
}

// SetSystemEnvVarToRegistry 设置系统环境变量（通过注册表）
func SetSystemEnvVarToRegistry(name, value string) error {
	// 打开系统环境变量注册表键
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, envRegistryPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %v", err)
	}
	defer key.Close()

	// 设置环境变量值
	if err := key.SetStringValue(name, value); err != nil {
		return fmt.Errorf("设置环境变量值失败: %v", err)
	}

	// 环境变量广播将在所有变量设置完成后统一执行一次

	return nil
}

// BroadcastEnvironmentChange 广播环境变量更改消息（导出函数）
func BroadcastEnvironmentChange() error {
	// 使用PowerShell脚本发送WM_SETTINGCHANGE消息，并使用try-catch捕获可能的错误
	psCmd := `
try {
    # 方法1: rundll32
    [void][System.Reflection.Assembly]::LoadWithPartialName('Microsoft.VisualBasic')
    [Microsoft.VisualBasic.Interaction]::Shell("rundll32 user32.dll,UpdatePerUserSystemParameters", 0, $false, 0)
    
    # 方法2: WM_SETTINGCHANGE
    $signature = @'
    [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
    public static extern IntPtr SendMessageTimeout(
        IntPtr hWnd, 
        uint Msg, 
        UIntPtr wParam, 
        string lParam,
        uint fuFlags, 
        uint uTimeout, 
        out UIntPtr lpdwResult);
'@
    $type = Add-Type -MemberDefinition $signature -Name "Win32SendMessage" -Namespace Win32Functions -PassThru
    $HWND_BROADCAST = [IntPtr]0xffff
    $WM_SETTINGCHANGE = 0x001A
    $result = [UIntPtr]::Zero
    [void]$type::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [UIntPtr]::Zero, "Environment", 2, 5000, [ref]$result)
    
    # 方法3: 临时修改TEMP变量
    $temp = [System.Environment]::GetEnvironmentVariable("TEMP", "Machine")
    if ($temp) {
        [System.Environment]::SetEnvironmentVariable("TEMP", $temp, "Machine")
    }
    
    Write-Host "SUCCESS: 环境变量已成功广播到系统" 
} catch {
    Write-Host "ERROR: 广播环境变量失败 - $($_.Exception.Message)"
    exit 1
}
`
	// 执行PowerShell脚本
	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil || strings.Contains(outputStr, "ERROR:") {
		return fmt.Errorf("广播环境变量失败: %v - %s", err, outputStr)
	}

	return nil
}
