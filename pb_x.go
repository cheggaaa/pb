// +build linux darwin freebsd openbsd solaris

package pb

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	TIOCGWINSZ     = 0x5413
	TIOCGWINSZ_OSX = 1074295912
)

// These constants are declared here, rather than importing
// them from the syscall package as some syscall packages, even
// on linux, for example gccgo, do not declare them.
const ioctlReadTermios = 0x5401  // syscall.TCGETS
const ioctlWriteTermios = 0x5402 // syscall.TCSETS

var tty *os.File

func init() {
	var err error
	tty, err = os.Open("/dev/tty")
	if err != nil {
		tty = os.Stdin
	}
}

func terminalWidth() (int, error) {
	w := new(window)
	tio := syscall.TIOCGWINSZ
	if runtime.GOOS == "darwin" {
		tio = TIOCGWINSZ_OSX
	}
	res, _, err := syscall.Syscall(sys_ioctl,
		tty.Fd(),
		uintptr(tio),
		uintptr(unsafe.Pointer(w)),
	)
	if int(res) == -1 {
		return 0, err
	}
	return int(w.Col), nil
}

var oldState syscall.Termios

func lockEcho() (quit chan int, err error) {
	fd := tty.Fd()
	if _, _, e := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); e != 0 {
		err = fmt.Errorf("Can't get terminal settings")
		return
	}

	newState := oldState
	newState.Lflag &^= syscall.ECHO
	newState.Lflag |= syscall.ICANON | syscall.ISIG
	newState.Iflag |= syscall.ICRNL
	if _, _, e := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlWriteTermios, uintptr(unsafe.Pointer(&newState)), 0, 0, 0); e != 0 {
		err = fmt.Errorf("Can't set terminal settings")
		return
	}
	quit = make(chan int, 1)
	go catchTerminate(quit)
	return
}

func unlockEcho() (err error) {
	fd := tty.Fd()
	if _, _, e := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlWriteTermios, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); e != 0 {
		err = fmt.Errorf("Can't set terminal settings")
	}
	return
}

// make like proxy for interupt signals for return terminal state
func catchTerminate(quit chan int) (err error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGKILL)
	defer signal.Stop(sig)

	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		return
	}

	for {
		select {
		case <-quit:
			unlockEcho()
			return
		case s := <-sig:
			unlockEcho()
			process.Signal(s)
			return
		}
	}
}
