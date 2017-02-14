package main

import "unsafe"

func main() {
	if IsMutexExists("Local\\{AB2F0A5E-FAA2-4664-B3C2-25D3984F0A20}") {
		return
	}

	AllocConsole()

	instanceHandle := GetCurrentInstance()
	className := "{8677407E-01E9-4D3E-8BF5-F9082CE08AEB}"
	windowName := "Monitor"
	wndExtra := uint32(unsafe.Sizeof(uintptr(0))) // reserve a space for self pointer

	CreateWindow(className, windowName, wndProc, WS_OVERLAPPEDWINDOW|WS_VISIBLE, instanceHandle, wndExtra)

	GetForegroundApp()

	MessageLoop()
}

func wndProc(hwnd HWND, msg uint32, wparam WPARAM, lparam LPARAM) LRESULT {
	switch msg {
	case WM_DESTROY:
		PostQuitMessage(0)
	default:
	}
	return DefWindowProc(hwnd, msg, wparam, lparam)
}
