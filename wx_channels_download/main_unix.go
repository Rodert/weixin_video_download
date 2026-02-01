//go:build !windows

package main

import (
	"fmt"

	"wx_channel/pkg/credit"
)

// disableConsoleQuickEdit 非 Windows 平台空实现
func disableConsoleQuickEdit() {}

// setConsoleTitle 非 Windows 平台空实现
func setConsoleTitle() {}

// setConsoleFont 非 Windows 平台空实现
func setConsoleFont() {}

// printVersion 打印版本信息（非 Windows 平台）
func printVersion() {
	// 读取 version.txt 中的版本号
	version := credit.GetCurrentVersion()
	
	versionInfo := fmt.Sprintf("视频号下载器\n版本: %s", version)
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println(versionInfo)
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("下载器下载地址: http://wx.wukongkt.vip:28088/")
	fmt.Println("联系微信: wangshiyu2046")
	fmt.Println("诚信质保")
	fmt.Println()
}

