package main

import (
	"flag"
	"time"
	"unsafe"
)

var (
	logger        *Logger
	timer         *time.Timer
	checkInterval = time.Duration(CHECK_INTERNVAL_IN_SECONDS) * time.Second
)

func main() {
	if IsMutexExists("Local\\{AB2F0A5E-FAA2-4664-B3C2-25D3984F0A20}") {
		return
	}

	verbose := flag.Bool("v", false, "open an extra window for debugging output")
	flag.Parse()

	var windowStyle uint32 = WS_OVERLAPPEDWINDOW
	if *verbose {
		OpenStdout() // useful for debugging output
		SetConsoleCtrlHandler(controlHandler)
		windowStyle |= WS_VISIBLE // somehow if !visible, the controlHandler won't be called
	}

	logger = NewLogger(CHECK_INTERNVAL_IN_SECONDS)
	defer logger.Close()

	instanceHandle := GetCurrentInstance()
	className := "{8677407E-01E9-4D3E-8BF5-F9082CE08AEB}"
	windowName := "Monitor"

	timer = time.NewTimer(checkInterval)
	hwnd := CreateWindow(className, windowName, wndProc, windowStyle, instanceHandle)
	registerNotifications(hwnd)

	go timerLoop()

	MessageLoop()
}

func timerLoop() {
	for {
		GetForegroundApp()
		<-timer.C
		startTimer()
	}
}

func wndProc(hwnd HWND, msg uint32, wparam WPARAM, lparam LPARAM) LRESULT {
	switch msg {

	case WM_POWERBROADCAST:
		if wparam == PBT_POWERSETTINGCHANGE && lparam != 0 {
			onPowerBroadcast((*POWERBROADCAST_SETTING)(unsafe.Pointer(lparam)))
		}

	case WM_WTSSESSION_CHANGE:
		onTerminalSessionChange(wparam)

	case WM_DESTROY:
		PostQuitMessage(0)

	default:
	}
	return DefWindowProc(hwnd, msg, wparam, lparam)
}

func registerNotifications(hwnd HWND) {
	RegisterPowerSettingNotification(hwnd, &GUID_SESSION_USER_PRESENCE, DEVICE_NOTIFY_WINDOW_HANDLE)
	RegisterPowerSettingNotification(hwnd, &GUID_SESSION_DISPLAY_STATUS, DEVICE_NOTIFY_WINDOW_HANDLE)
	WTSRegisterSessionNotification(hwnd, NOTIFY_FOR_THIS_SESSION)
}

func onPowerBroadcast(setting *POWERBROADCAST_SETTING) {
	if setting.PowerSetting == GUID_SESSION_USER_PRESENCE {
		const userPresnet = 0
		const userInactive = 2
		data := setting.Data
		if data == userInactive {
			stopTimer()
		} else if data == userPresnet {
			startTimer()
		}
	} else if setting.PowerSetting == GUID_SESSION_DISPLAY_STATUS {
		const displayOff = 0
		const displayOn = 1
		data := setting.Data
		if data == displayOff {
			stopTimer()
		} else if data == displayOn {
			startTimer()
		}
	}
}

func onTerminalSessionChange(data WPARAM) {
	const sessionLock = 7
	const sessionUnlock = 8
	if data == sessionLock {
		stopTimer()
	} else if data == sessionUnlock {
		startTimer()
	}
}

func stopTimer() {
	if !timer.Stop() {
		// NOTE: document calls for drain the channel right after stop, but that actually blocks
		// my further execution. remove the drain for now
		//
		// Original code:
		// <-timer.C
	}
}

func startTimer() {
	// NOTE: document calls for stop timer first, however go\src\time\sleep.go!Reset calls stopTimer(),
	// thus I have no reason to double stop it here
	//
	// Original code:
	// stopTimer()

	timer.Reset(checkInterval)
}

func controlHandler(ctrlType DWORD) uintptr {
	if ctrlType == CTRL_CLOSE_EVENT {
		logger.Log("Console Closed", "Console Closed")
		logger.Close()
		// ExitThread(0)
		// return TRUE
	}
	return FALSE
}
