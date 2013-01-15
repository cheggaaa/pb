// +build windows

package pb

import (
	"syscall"

	"github.com/AllenDang/w32"
)

func bold(str string) string {
	return str
}

func terminalWidth() (int, error) {
	screenBufInfo := w32.GetConsoleScreenBufferInfo(w32.HANDLE(syscall.Stdout))
	if screenBufInfo == nil {
		return 79, nil
	}
	return int(screenBufInfo.DwSize.X) - 1, nil
}
