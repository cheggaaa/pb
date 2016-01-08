// +build darwin freebsd netbsd openbsd solaris dragonfly

package pb

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA
const ioctlWriteTermios = syscall.TIOCSETA
