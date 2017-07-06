package main

import (
	"os"
	"syscall"
	"unsafe"
	"strconv"
)

// Terminal ...
type Terminal struct {
	Width			int
	Height			int
}

// NewTerminal ...
func NewTerminal() *Terminal {
	width, height := 0, 0

	s1, ok1 := os.LookupEnv("COLUMNS")
	s2, ok2 := os.LookupEnv("LINES")

	if ok1 && ok2 {
		width, _ = strconv.Atoi(s1)
		height, _ = strconv.Atoi(s2)
	} else {
		win := &struct { h, w, x, y uint16 }{}
		code, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin), uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(win)))
		if code == 0 {
			width = int(win.w)
			height = int(win.h)
		}
	}

	return &Terminal{
		Width:			width,
		Height:			height,
	}
}
