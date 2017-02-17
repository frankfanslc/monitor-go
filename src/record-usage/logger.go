package main

import (
	"fmt"
	"os"
	"time"
)

const MAX_ENTRIES_BEFORE_FLUSH = 3

type logEntry struct {
	timestamp         time.Time
	durationInSeconds uint32
	windowTitle       string
	commandLine       string
}

type Logger struct {
	file              *os.File
	intervalInSeconds uint32
	count             uint32
	entries           [MAX_ENTRIES_BEFORE_FLUSH]logEntry
}

func NewLogger(intervalInSeconds uint32) *Logger {
	filename := os.Getenv("LOCALAPPDATA") + "/record-usage.csv"
	fmt.Println(filename)
	file, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0)
	return &Logger{file: file, intervalInSeconds: intervalInSeconds}
}

func (x *Logger) Close() {
	if x.file == nil {
		return
	}
	x.flush()
	x.file.Close()
	x.file = nil
}

func (x *Logger) Log(windowsTitle string, commandLine string) {
	if x.file == nil {
		return
	}
	entry := logEntry{
		timestamp:         time.Now(),
		durationInSeconds: x.intervalInSeconds,
		windowTitle:       windowsTitle,
		commandLine:       commandLine,
	}

	// case 1 - completely empty
	if x.count == 0 {
		x.entries[0] = entry
		x.count++
		return
	}

	// case 2 - same event as last one
	lastEntry := &x.entries[x.count-1]
	if lastEntry.windowTitle == windowsTitle && lastEntry.commandLine == commandLine {
		lastEntry.durationInSeconds += x.intervalInSeconds
		return
	}

	// case 3 - need flush first
	if x.count == MAX_ENTRIES_BEFORE_FLUSH {
		x.flush()
	}

	// case 4 - add the new entry
	x.entries[x.count] = entry
	x.count++
}

func (x *Logger) flush() {
	var i uint32
	for i = 0; i < x.count; i++ {
		entry := x.entries[i]
		// RFC3339 = "2006-01-02T15:04:05Z07:00"
		timestamp := entry.timestamp.Format(time.RFC3339)
		fmt.Fprintf(x.file, "%s, %4d, %s, %s\n",
			timestamp,
			entry.durationInSeconds,
			entry.commandLine,
			entry.windowTitle)
	}
	x.count = 0
}
