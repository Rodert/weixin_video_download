package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"wx_channel/cmd"
	"wx_channel/config"
	"wx_channel/internal/interceptor"
	"wx_channel/pkg/credit"
	"wx_channel/pkg/platform"
)

//go:embed certs/SunnyRoot.cer
var cert_file []byte

//go:embed certs/private.key
var private_key_file []byte

//go:embed inject/lib/FileSaver.min.js
var js_file_saver []byte

//go:embed inject/lib/jszip.min.js
var js_zip []byte

//go:embed inject/lib/recorder.min.js
var js_recorder []byte

//go:embed inject/pagespy.min.js
var js_pagespy []byte

//go:embed inject/pagespy.js
var js_debug []byte

//go:embed inject/error.js
var js_error []byte

//go:embed inject/utils.js
var js_utils []byte

//go:embed inject/main.js
var js_main []byte

//go:embed inject/live.js
var js_live_main []byte

//go:embed inject/download_list.js
var js_download_list []byte

//go:embed version.txt
var embeddedVersion []byte

var FilesCert = &interceptor.ServerCertFiles{
	CertFile:       cert_file,
	PrivateKeyFile: private_key_file,
}
var FilesChannelScript = &interceptor.ChannelInjectedFiles{
	JSFileSaver:    js_file_saver,
	JSZip:          js_zip,
	JSRecorder:     js_recorder,
	JSPageSpy:      js_pagespy,
	JSDebug:        js_debug,
	JSError:        js_error,
	JSUtils:        js_utils,
	JSMain:         js_main,
	JSLiveMain:     js_live_main,
	JSDownloadList: js_download_list,
}

var RootCertificateName = "SunnyNet"

// EnableLogs 是否启用日志打印（构建时通过 ldflags 传入，默认 false）
var EnableLogs = "false"

// isDevMode 检测是否是开发模式（go run 运行）或启用了日志
func isDevMode() bool {
	// 如果构建时设置了 EnableLogs=true，则启用日志
	if EnableLogs == "true" {
		return true
	}

	// 如果构建时明确设置了 EnableLogs=false，则禁用日志（生产模式）
	// 注意：默认值 "false" 也会走这里，所以打包后的 exe 默认不打印日志
	if EnableLogs == "false" {
		return false
	}

	// EnableLogs 既不是 "true" 也不是 "false"（默认值或未设置）
	// 检查是否是开发模式（go run）
	exe, err := os.Executable()
	if err != nil {
		return true // 如果获取失败，默认认为是开发模式
	}
	exeLower := strings.ToLower(exe)

	// 检查路径中是否包含 go-build（go run 的临时目录）
	// 或者可执行文件名是 main.exe（开发时常见）
	// 或者路径包含临时目录标识
	isDev := strings.Contains(exeLower, "go-build") ||
		strings.Contains(exeLower, "main.exe") ||
		strings.Contains(exeLower, "\\temp\\") ||
		strings.Contains(exeLower, "/tmp/")

	return isDev
}

// getBuildTime 获取当前日期（精确到天）
func getBuildTime() string {
	return time.Now().Format("2006-01-02")
}

// getVersion 获取版本号（使用构建日期）
func getVersion() string {
	return getBuildTime()
}

// getEmbeddedVersionText 获取嵌入的版本号文本
func getEmbeddedVersionText() string {
	if len(embeddedVersion) == 0 {
		return ""
	}
	// 去除所有空白字符（空格、制表符、换行符、回车符等）
	version := strings.TrimSpace(string(embeddedVersion))
	// 去除 BOM（字节顺序标记）
	version = strings.TrimPrefix(version, "\ufeff")
	// 去除所有空白字符后再次检查
	version = strings.TrimSpace(version)
	// 只取第一行（如果有换行符）
	if idx := strings.Index(version, "\n"); idx > 0 {
		version = strings.TrimSpace(version[:idx])
	}
	if idx := strings.Index(version, "\r"); idx > 0 {
		version = strings.TrimSpace(version[:idx])
	}
	return version
}

func main() {
	// 设置嵌入版本号的获取函数（优先使用嵌入的版本号）
	credit.SetEmbeddedVersionFunc(getEmbeddedVersionText)
	
	// Windows 特定处理：设置控制台标题和禁用快速编辑模式
	setConsoleTitle()
	setConsoleFont()
	printVersion()
	disableConsoleQuickEdit()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("加载配置文件失败 %v", err.Error())
		return
	}
	if cfg.ProxySystem && platform.NeedAdminPermission() && !platform.IsAdmin() {
		if !platform.RequestAdminPermission() {
			fmt.Println("启动失败，请右键选择「以管理员身份运行」")
			return
		}
		return
	}
	if err := cmd.Execute(getVersion(), RootCertificateName, FilesChannelScript, FilesCert, cfg, isDevMode()); err != nil {
		fmt.Printf("初始化失败 %v\n", err.Error())
	}
}

