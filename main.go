package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"switch/config"
	"switch/jdk"
)

const version = "1.0.0"

func showHelp() {
	fmt.Println("JDK Switch Tool v" + version)
	fmt.Println("用法: jdk-switch [命令]")
	fmt.Println("\n命令:")
	fmt.Println("  -init      初始化配置文件")
	fmt.Println("  -list      列出所有可用的JDK版本")
	fmt.Println("  -set <版本> 切换到指定的JDK版本")
	fmt.Println("  -backup    仅备份当前环境变量，不切换JDK版本")
	fmt.Println("  -v         显示版本信息")
	fmt.Println("  -h         显示帮助信息")
	fmt.Println("\n不带参数运行将启动交互模式")
	fmt.Println("\n环境变量备份信息:")
	fmt.Println("  每次切换JDK版本时会自动备份当前的环境变量(PATH, JAVA_HOME, CLASSPATH)")
	fmt.Println("  备份文件存储位置: C:\\jdk-switch\\backup\\时间戳")
	fmt.Println("\n提示:")
	fmt.Println("  切换JDK版本后，重新打开命令行窗口或重新登录系统，以确保新的Java版本生效")
}

// 询问用户是否要初始化配置
func askForInit() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("是否要初始化配置文件？(y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		}
		fmt.Println("请输入 y 或 n")
	}
}

func main() {
	// 解析命令行参数
	initFlag := flag.Bool("init", false, "初始化配置文件")
	listFlag := flag.Bool("list", false, "列出所有可用的JDK版本")
	setVersion := flag.String("set", "", "切换到指定的JDK版本")
	backupFlag := flag.Bool("backup", false, "仅备份当前环境变量，不切换JDK版本")
	versionFlag := flag.Bool("v", false, "显示版本信息")
	helpFlag := flag.Bool("h", false, "显示帮助信息")
	flag.Parse()

	// 显示版本信息
	if *versionFlag {
		fmt.Println("JDK Switch Tool v" + version)
		return
	}

	// 显示帮助信息
	if *helpFlag {
		showHelp()
		return
	}

	// 如果是备份环境变量命令
	if *backupFlag {
		if err := jdk.BackupEnvironmentVariables(); err != nil {
			fmt.Printf("备份环境变量失败: %v\n", err)
			return
		}
		return
	}

	// 如果是初始化命令
	if *initFlag {
		if err := config.InitDefaultConfig(); err != nil {
			fmt.Printf("初始化配置失败: %v\n", err)
			return
		}
		fmt.Printf("配置文件已初始化，路径: %s\\%s\n", config.DefaultDir, config.DefaultFile)
		fmt.Println("请根据实际情况修改JDK路径")
		return
	}

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("未找到配置文件: %s\\%s\n", config.DefaultDir, config.DefaultFile)

			// 询问用户是否要初始化配置
			if askForInit() {
				if err := config.InitDefaultConfig(); err != nil {
					fmt.Printf("初始化配置失败: %v\n", err)
					return
				}
				fmt.Printf("配置文件已初始化，路径: %s\\%s\n", config.DefaultDir, config.DefaultFile)
				fmt.Println("请根据实际情况修改JDK路径后重新运行程序")
				return
			}
			fmt.Println("您可以稍后使用 -init 参数初始化配置")
			return
		}
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 列出所有JDK版本
	if *listFlag {
		fmt.Printf("当前JDK版本: %s\n", cfg.CurrentVersion)
		fmt.Println("可用的JDK版本:")
		for version := range cfg.JDKPaths {
			if version == cfg.CurrentVersion {
				fmt.Printf("* JDK %s: %s (当前)\n", version, cfg.JDKPaths[version])
			} else {
				fmt.Printf("  JDK %s: %s\n", version, cfg.JDKPaths[version])
			}
		}
		return
	}

	// 切换到指定版本
	if *setVersion != "" {
		if err := switchJDK(cfg, *setVersion); err != nil {
			fmt.Printf("错误: %v\n", err)
			return
		}
		fmt.Printf("成功切换到JDK %s\n", *setVersion)
		return
	}

	// 交互模式
	fmt.Println("JDK Switch Tool v" + version)
	fmt.Printf("当前JDK版本: %s\n", cfg.CurrentVersion)
	fmt.Println("可用的JDK版本:")
	for version := range cfg.JDKPaths {
		if version == cfg.CurrentVersion {
			fmt.Printf("* JDK %s: %s (当前)\n", version, cfg.JDKPaths[version])
		} else {
			fmt.Printf("  JDK %s: %s\n", version, cfg.JDKPaths[version])
		}
	}

	// 读取用户输入
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n请输入要切换的JDK版本 (输入 'b' 备份环境变量, 输入 'q' 退出): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" {
			break
		}

		if input == "b" {
			if err := jdk.BackupEnvironmentVariables(); err != nil {
				fmt.Printf("备份环境变量失败: %v\n", err)
				continue
			}
			continue
		}

		if err := switchJDK(cfg, input); err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}

		fmt.Printf("成功切换到JDK %s\n", input)
	}
}

// 切换JDK版本的通用函数
func switchJDK(cfg *config.Config, version string) error {
	// 获取对应的JDK路径
	jdkPath, err := cfg.GetJDKPath(version)
	if err != nil {
		return err
	}

	// 验证JDK路径
	if !jdk.ValidateJDKPath(jdkPath) {
		return fmt.Errorf("无效的JDK路径 - %s", jdkPath)
	}

	// 切换JDK
	if err := jdk.SetJavaHome(jdkPath); err != nil {
		return fmt.Errorf("切换JDK失败: %v", err)
	}

	// 更新当前版本
	if err := cfg.UpdateCurrentVersion(version); err != nil {
		return fmt.Errorf("更新配置失败: %v", err)
	}

	// 保存配置
	if err := cfg.SaveConfig(); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	// 添加简洁明确的提示信息
	fmt.Println("\n环境变量已成功更新。如需使用新的Java版本，请:")
	fmt.Println("- 重新打开一个新的命令行窗口")
	fmt.Println("- 或使用 refreshenv 命令（如果安装了Chocolatey）")

	return nil
}
