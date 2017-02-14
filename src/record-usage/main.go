package main

import "time"
import "fmt"

func GetForegroundApp() {
	seconds := 3
	time.Sleep(time.Duration(seconds) * time.Second)

	hwnd := GetForegroundWindow()
	if hwnd == 0 {
		return
	}

	processID := GetWindowProcessId(hwnd)
	fmt.Println("pid: ", processID)

	processHandle := OpenProcess(processID)
	if processHandle == 0 {
		return
	}

	windowText := GetWindowText(hwnd)
	fmt.Println("title: ", windowText)

	if IsImmersiveProcess(processHandle) {
		CloseHandle(processHandle)
		hwnd, processID = GetUniversalApp(hwnd, processID)
		processHandle = OpenProcess(processID)
		fmt.Println("pid: ", processID)
	}

	commandLine := GetProcessCommandLine(processHandle)
	fmt.Println("cmd: ", commandLine)
}
