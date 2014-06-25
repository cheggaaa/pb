// +build !windows

package pb

import (
	"runtime"
	"syscall"
	"unsafe"
)

const (
	TIOCGWINSZ     = 0x5413
	TIOCGWINSZ_OSX = 1074295912

	SYS_IOCTL_SOLARIS = 54
)

func bold(str string) string {
	return "\033[1m" + str + "\033[0m"
}

func terminalWidth() (int, error) {
	w := new(window)
	tio := syscall.TIOCGWINSZ
	sys_ioctl := syscall.SYS_IOCTL
	switch runtime.GOOS {
	case "darwin":
		tio = TIOCGWINSZ_OSX
		break
	case "solaris":
		sys_ioctl = SYS_IOCTL_SOLARIS
		break
	}

	res, _, err := syscall.Syscall(uintptr(sys_ioctl),
		uintptr(syscall.Stdin),
		uintptr(tio),
		uintptr(unsafe.Pointer(w)),
	)
	if int(res) == -1 {
		return 0, err
	}
	return int(w.Col), nil
}
