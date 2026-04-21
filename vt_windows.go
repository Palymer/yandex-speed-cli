//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

// tryEnableVirtualTerminal — ANSI (в т.ч. \033[2K) в стандартной консоли Windows 10+.
func tryEnableVirtualTerminal() {
	h := os.Stdout.Fd()
	var mode uint32
	r, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return
	}
	const enableVirtualTerminalProcessing = 0x0004
	if mode&enableVirtualTerminalProcessing != 0 {
		return
	}
	procSetConsoleMode.Call(h, uintptr(mode|enableVirtualTerminalProcessing))
}
