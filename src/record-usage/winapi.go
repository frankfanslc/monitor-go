package main

import "unsafe"
import "syscall"

type (
	BOOL      int32
	DWORD     uint32
	HANDLE    uintptr
	HINSTANCE HANDLE
	HWND      HANDLE
	WPARAM    uintptr
	LPARAM    uintptr
	LRESULT   uintptr
	WCHAR     uint16
	PWSTR     uintptr
	CALLBACK  uintptr
)

const (
	FALSE = 0
	TRUE  = 1
)

const (
	WM_DESTROY = 0x0002
	WM_CLOSE   = 0x0010

	COLOR_BACKGROUND = 1

	CW_USEDEFAULT = 0x80000000
)

const (
	WS_OVERLAPPED       = 0x00000000
	WS_POPUP            = 0x80000000
	WS_CHILD            = 0x40000000
	WS_MINIMIZE         = 0x20000000
	WS_VISIBLE          = 0x10000000
	WS_DISABLED         = 0x08000000
	WS_CLIPSIBLINGS     = 0x04000000
	WS_CLIPCHILDREN     = 0x02000000
	WS_MAXIMIZE         = 0x01000000
	WS_CAPTION          = 0x00C00000
	WS_BORDER           = 0x00800000
	WS_DLGFRAME         = 0x00400000
	WS_VSCROLL          = 0x00200000
	WS_HSCROLL          = 0x00100000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_GROUP            = 0x00020000
	WS_TABSTOP          = 0x00010000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_TILED            = WS_OVERLAPPED
	WS_ICONIC           = WS_MINIMIZE
	WS_SIZEBOX          = WS_THICKFRAME
	WS_TILEDWINDOW      = WS_OVERLAPPEDWINDOW
	WS_OVERLAPPEDWINDOW = (WS_OVERLAPPED |
		WS_CAPTION |
		WS_SYSMENU |
		WS_THICKFRAME |
		WS_MINIMIZEBOX |
		WS_MAXIMIZEBOX)
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	ntdll    = syscall.NewLazyDLL("ntdll.dll")

	procDefWindowProc = user32.NewProc("DefWindowProcW")
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
	return syscall.UTF16ToString(buffer[:])
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

func ReadProcessMemory(process HANDLE, readAddress uintptr, writeBuffer unsafe.Pointer, size uintptr) bool {
	var bytesRead uintptr
	result, _, _ := kernel32.NewProc("ReadProcessMemory").Call(uintptr(process), readAddress, uintptr(writeBuffer), size, uintptr(unsafe.Pointer(&bytesRead)))
	return result != 0
}

const (
	ProcessBasicInformation = 0
	ProcessWow64Information = 26
)

type POINTER32 uint32

type PROCESS_BASIC_INFORMATION struct {
	Reserved1       uintptr
	PebBaseAddress  uintptr
	Reserved2       [2]uintptr
	UniqueProcessId uintptr
	Reserved3       uintptr
}

type PROCESS_ENVIRONMENT_BLOCK struct {
	Reserved12        uintptr
	Reserved3         [2]uintptr
	Ldr               uintptr
	ProcessParameters uintptr
}

type PROCESS_ENVIRONMENT_BLOCK32 struct {
	Reserved12        POINTER32
	Reserved3         [2]POINTER32
	Ldr               POINTER32
	ProcessParameters POINTER32
}

type UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Pad1          uint32
	Buffer        uintptr
}

type UNICODE_STRING32 struct {
	Length        uint16
	MaximumLength uint16
	Buffer        POINTER32
}

type RTL_USER_PROCESS_PARAMETERS struct {
	Reserved1     [16]uint8
	Resserved2    [10]uintptr
	ImagePathName UNICODE_STRING
	CommandLine   UNICODE_STRING
}

type RTL_USER_PROCESS_PARAMETERS32 struct {
	Reserved1     [16]uint8
	Resserved2    [10]POINTER32
	ImagePathName UNICODE_STRING32
	CommandLine   UNICODE_STRING32
}

func getProcessPebAddress(process HANDLE) uintptr {
	var buffer PROCESS_BASIC_INFORMATION
	var returnLength uintptr
	ntdll.NewProc("NtQueryInformationProcess").Call(
		uintptr(process),
		ProcessBasicInformation,
		uintptr(unsafe.Pointer(&buffer)),
		unsafe.Sizeof(buffer),
		uintptr(unsafe.Pointer(&returnLength)))
	return buffer.PebBaseAddress
}

func getProcessPebAddress32(process HANDLE) uintptr {
	var buffer uintptr
	var returnLength uintptr
	ntdll.NewProc("NtQueryInformationProcess").Call(
		uintptr(process),
		ProcessWow64Information,
		uintptr(unsafe.Pointer(&buffer)),
		unsafe.Sizeof(buffer),
		uintptr(unsafe.Pointer(&returnLength)))
	return buffer
}

func IsWow64Process(process HANDLE) bool {
	var result uintptr
	kernel32.NewProc("IsWow64Process").Call(uintptr(process), uintptr(unsafe.Pointer(&result)))
	return result != FALSE
}

func GetProcessCommandLine(process HANDLE) string {
	if IsWow64Process(process) {
		return GetProcessCommandLine32(process)
	}

	pebAddress := getProcessPebAddress(process)

	var peb PROCESS_ENVIRONMENT_BLOCK
	ReadProcessMemory(process, pebAddress, unsafe.Pointer(&peb), unsafe.Sizeof(peb))

	var userProcessParameters RTL_USER_PROCESS_PARAMETERS
	ReadProcessMemory(process, peb.ProcessParameters, unsafe.Pointer(&userProcessParameters), unsafe.Sizeof(userProcessParameters))

	byteCount := userProcessParameters.CommandLine.Length
	charCount := byteCount / 2
	var buffer [256]uint16
	if charCount > 256 {
		byteCount = 256 * 2
	}
	ReadProcessMemory(process, userProcessParameters.CommandLine.Buffer, unsafe.Pointer(&buffer), uintptr(byteCount))
	return syscall.UTF16ToString(buffer[:])
}

func GetProcessCommandLine32(process HANDLE) string {

	pebAddress := getProcessPebAddress32(process)

	var peb PROCESS_ENVIRONMENT_BLOCK32
	ReadProcessMemory(process, pebAddress, unsafe.Pointer(&peb), unsafe.Sizeof(peb))

	var userProcessParameters RTL_USER_PROCESS_PARAMETERS32
	ReadProcessMemory(process, uintptr(peb.ProcessParameters), unsafe.Pointer(&userProcessParameters), unsafe.Sizeof(userProcessParameters))

	byteCount := userProcessParameters.CommandLine.Length
	charCount := byteCount / 2
	var buffer [256]uint16
	if charCount > 256 {
		byteCount = 256 * 2
	}
	ReadProcessMemory(process, uintptr(userProcessParameters.CommandLine.Buffer), unsafe.Pointer(&buffer), uintptr(byteCount))
	return syscall.UTF16ToString(buffer[:])
}

func stringToPVoid(s string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}

func CreateMutex(name string) HANDLE {
	initialOwner := FALSE
	result, _, _ := kernel32.NewProc("CreateMutexW").Call(uintptr(0), uintptr(initialOwner), stringToPVoid(name))
	return HANDLE(result)
}

func IsMutexExists(name string) bool {
	initialOwner := FALSE
	result, _, err := kernel32.NewProc("CreateMutexW").Call(uintptr(0), uintptr(initialOwner), stringToPVoid(name))
	return result == uintptr(0) || err == syscall.ERROR_ALREADY_EXISTS
}

func AllocConsole() {
	kernel32.NewProc("AllocConsole").Call()
}

type WNDCLASS struct {
	style         uint32
	lpfnWndProc   CALLBACK
	cbClsExtra    uint32
	cbWndExtra    uint32
	hInstance     HINSTANCE
	hIcon         HANDLE
	hCursor       HANDLE
	hbrBackground HANDLE
	lpszMenuName  PWSTR
	lpszClassName PWSTR
}

type POINT struct {
	x int32
	y int32
}

type MSG struct {
	hwnd    HWND
	message uint32
	wParam  WPARAM
	lParam  LPARAM
	time    uint32
	pt      POINT
}

func DefWindowProc(hwnd HWND, msg uint32, wparam WPARAM, lparam LPARAM) LRESULT {
	result, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(msg), uintptr(wparam), uintptr(lparam))
	return LRESULT(result)
}

func PostQuitMessage(exitCode int32) {
	user32.NewProc("PostQuitMessage").Call(uintptr(exitCode))
}

func CreateWindow(className string, windowName string, wndProc interface{}, windowStyle uint32, instance HINSTANCE, wndExtra uint32) HWND {
	wndProcCallback := syscall.NewCallback(wndProc)
	wndclass := WNDCLASS{
		lpfnWndProc:   CALLBACK(wndProcCallback),
		cbWndExtra:    wndExtra,
		hbrBackground: HANDLE(COLOR_BACKGROUND),
		lpszClassName: PWSTR(stringToPVoid(className)),
	}

	result, _, _ := user32.NewProc("RegisterClassW").Call(uintptr(unsafe.Pointer(&wndclass)))
	if result == 0 {
		return HWND(0)
	}

	result, _, _ = user32.NewProc("CreateWindowExW").Call(
		uintptr(0), // extStyle
		stringToPVoid(className),
		stringToPVoid(windowName),
		uintptr(windowStyle),
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		uintptr(0), // hwndParent
		uintptr(0), // hMenu
		uintptr(instance),
		wndProcCallback) // Passed to WM_NCCREATE as CREATESTRUCT.lpCreateParams

	return HWND(result)
}

func GetModuleHandle(name string) uintptr {
	result, _, _ := kernel32.NewProc("GetModuleHandleW").Call(stringToPVoid(name))
	return result
}

func GetCurrentInstance() HINSTANCE {
	result, _, _ := kernel32.NewProc("GetModuleHandleW").Call(uintptr(0))
	return HINSTANCE(result)
}

func MessageLoop() {
	var msg MSG
	msgPtr := uintptr(unsafe.Pointer(&msg))
	procGetMessage := user32.NewProc("GetMessageW")
	procDispatchMessage := user32.NewProc("DispatchMessageW")
	for {
		result, _, _ := procGetMessage.Call(msgPtr, uintptr(0), uintptr(0), uintptr(0))
		if result == uintptr(FALSE) {
			break
		}
		procDispatchMessage.Call(msgPtr)
	}
}
