package main

import "unsafe"
import "golang.org/x/sys/windows"

type (
	BOOL      int32
	DWORD     uint32
	HANDLE    uintptr
	HINSTANCE HANDLE
	HWND      HANDLE
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	user32   = windows.NewLazySystemDLL("user32.dll")
)

func GetForegroundWindow() HWND {
	hwnd, _, _ := user32.NewProc("GetForegroundWindow").Call()
	return HWND(hwnd)
}

func GetWindowThreadProcessId(hwnd HWND) DWORD {
	var processID DWORD
	user32.NewProc("GetWindowThreadProcessId").Call(uintptr(hwnd), uintptr(unsafe.Pointer(&processID)))
	return DWORD(processID)
}
