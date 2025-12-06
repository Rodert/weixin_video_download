package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"wx_channel/cmd"
	"wx_channel/config"
	"wx_channel/internal/interceptor"
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

func main() {
	// Windows 特定处理：设置控制台标题和禁用快速编辑模式
	if runtime.GOOS == "windows" {
		setConsoleTitle()
		setConsoleFont()
		printVersion()
		disableConsoleQuickEdit()
	}

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

// disableConsoleQuickEdit 禁用 Windows 控制台的快速编辑模式
// 这可以避免双击运行 .exe 时控制台卡住等待输入的问题
func disableConsoleQuickEdit() {
	if runtime.GOOS != "windows" {
		return
	}

	// 尝试获取控制台句柄
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	var mode uint32
	consoleHandle := syscall.Handle(os.Stdin.Fd())

	// 获取当前模式
	ret, _, _ := getConsoleMode.Call(uintptr(consoleHandle), uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return // 获取失败，可能不是控制台环境
	}

	// 禁用快速编辑模式 (ENABLE_QUICK_EDIT_MODE = 0x0040)
	// 禁用插入模式 (ENABLE_INSERT_MODE = 0x0020)
	mode &^= 0x0040 | 0x0020

	// 设置新模式
	setConsoleMode.Call(uintptr(consoleHandle), uintptr(mode))
}

// setConsoleTitle 设置 Windows 控制台窗口标题
func setConsoleTitle() {
	if runtime.GOOS != "windows" {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleTitle := kernel32.NewProc("SetConsoleTitleW")

	title := fmt.Sprintf("视频号下载器 - 版本: %s", getBuildTime())
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	setConsoleTitle.Call(uintptr(unsafe.Pointer(titlePtr)))
}

// setConsoleFont 设置控制台字体大小
func setConsoleFont() {
	if runtime.GOOS != "windows" {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	setCurrentConsoleFontEx := kernel32.NewProc("SetCurrentConsoleFontEx")

	// STD_OUTPUT_HANDLE = -11
	stdoutHandle, _, _ := getStdHandle.Call(uintptr(0xFFFFFFF5)) // -11

	// CONSOLE_FONT_INFOEX 结构体
	type CONSOLE_FONT_INFOEX struct {
		cbSize      uint32
		nFont       uint32
		dwFontSizeX int16
		dwFontSizeY int16
		fontFamily  uint32
		fontWeight  uint32
		faceName    [32]uint16
	}

	fontInfo := CONSOLE_FONT_INFOEX{}
	fontInfo.cbSize = uint32(unsafe.Sizeof(fontInfo))
	fontInfo.dwFontSizeX = 0  // 宽度自动
	fontInfo.dwFontSizeY = 20 // 高度 20（大字体）
	fontInfo.fontFamily = 54  // FF_DONTCARE
	fontInfo.fontWeight = 400 // FW_NORMAL

	setCurrentConsoleFontEx.Call(stdoutHandle, 0, uintptr(unsafe.Pointer(&fontInfo)))
}

// printVersion 打印版本信息（大字体）
func printVersion() {
	if runtime.GOOS != "windows" {
		return
	}

	version := fmt.Sprintf("视频号下载器\n版本: %s", getBuildTime())
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println(version)
	fmt.Println("========================================")
	fmt.Println()
}
