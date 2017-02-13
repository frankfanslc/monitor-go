package main

import "unsafe"
import "golang.org/x/sys/windows"
import "syscall"

type (
	BOOL      int32
	DWORD     uint32
	HANDLE    uintptr
	HINSTANCE HANDLE
	HWND      HANDLE
	WPARAM    uint64
	LPARAM    int64
	WCHAR     uint16
)

const (
	FALSE = 0
	TRUE  = 1
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	user32   = windows.NewLazySystemDLL("user32.dll")
)

func GetForegroundWindow() HWND {
	hwnd, _, _ := user32.NewProc("GetForegroundWindow").Call()
	return HWND(hwnd)
}

func GetWindowProcessId(hwnd HWND) DWORD {
	var processID DWORD
	user32.NewProc("GetWindowThreadProcessId").Call(uintptr(hwnd), uintptr(unsafe.Pointer(&processID)))
	return DWORD(processID)
}

func OpenProcess(processID DWORD) HANDLE {
	const PROCESS_QUERY_INFORMATION = 0x00000400
	const PROCESS_VM_READ = 0x0010

	handle, _, _ := kernel32.NewProc("OpenProcess").Call(
		PROCESS_QUERY_INFORMATION|PROCESS_VM_READ,
		FALSE,
		uintptr(processID))
	return HANDLE(handle)
}

func CloseHandle(handle HANDLE) {
	kernel32.NewProc("CloseHandle").Call(uintptr(handle))
}

func GetWindowText(hwnd HWND) string {
	var buffer [256]uint16
	user32.NewProc("GetWindowTextW").Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buffer)), 256)
	return windows.UTF16ToString(buffer[:])
}

func IsImmersiveProcess(handle HANDLE) bool {
	result, _, _ := user32.NewProc("IsImmersiveProcess").Call(uintptr(handle))
	return result != FALSE
}

func GetUniversalApp(hwnd HWND, processID DWORD) (HWND, DWORD) {

	type enumParameter struct {
		hwnd           HWND
		processID      DWORD
		childHwnd      HWND
		childProcessID DWORD
	}
	parameter := enumParameter{hwnd: hwnd, processID: processID}

	callback := syscall.NewCallback(
		// Use uintptr instead of BOOL, because: syscall.NewCallback requires
		// the callback's returns size = unsafe.Sizeof(uintptr(0))
		// ref: \go\src\runtime\syscall_windows.go, syscall.compileCallback()
		func(childHwnd HWND, lparam LPARAM) uintptr {
			parameter := (*enumParameter)(unsafe.Pointer(uintptr(lparam)))
			childProcess := GetWindowProcessId(childHwnd)
			if childProcess != 0 && childProcess != parameter.processID {
				parameter.childHwnd = childHwnd
				parameter.childProcessID = childProcess
				return FALSE
			}
			return TRUE
		})

	user32.NewProc("EnumChildWindows").Call(
		uintptr(hwnd),
		callback,
		uintptr(unsafe.Pointer(&parameter)))

	return parameter.childHwnd, parameter.childProcessID
}
