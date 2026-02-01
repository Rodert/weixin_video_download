//go:build !windows

package main

// disableConsoleQuickEdit 非 Windows 平台空实现
func disableConsoleQuickEdit() {}

// setConsoleTitle 非 Windows 平台空实现
func setConsoleTitle() {}

// setConsoleFont 非 Windows 平台空实现
func setConsoleFont() {}

// printVersion 非 Windows 平台空实现
func printVersion() {}

