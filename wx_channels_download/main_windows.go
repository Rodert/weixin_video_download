//go:build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// disableConsoleQuickEdit 禁用 Windows 控制台的快速编辑模式
// 这可以避免双击运行 .exe 时控制台卡住等待输入的问题
func disableConsoleQuickEdit() {
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
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleTitle := kernel32.NewProc("SetConsoleTitleW")

	title := fmt.Sprintf("视频号下载器 - 版本: %s", getBuildTime())
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	setConsoleTitle.Call(uintptr(unsafe.Pointer(titlePtr)))
}

// setConsoleFont 设置控制台字体大小
func setConsoleFont() {
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
	version := fmt.Sprintf("视频号下载器\n版本: %s", getBuildTime())
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println(version)
	fmt.Println("========================================")
	fmt.Println()
}
