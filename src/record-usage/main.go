package main

import "time"
import "fmt"

func main() {
	seconds := 3
	time.Sleep(time.Duration(seconds) * time.Second)

	hwnd := GetForegroundWindow()
	if hwnd == 0 {
		return
	}

	processID := GetWindowThreadProcessId(hwnd)
	fmt.Println(processID)
}
