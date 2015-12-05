// +build linux darwin freebsd netbsd openbsd dragonfly

package pb

import "syscall"

const sys_ioctl = syscall.SYS_IOCTL
