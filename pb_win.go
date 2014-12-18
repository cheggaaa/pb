// +build windows

package pb

import (
	"github.com/olekukonko/ts"
)

var tty = os.Stdin

func terminalWidth() (int, error) {
	size, err := ts.GetSize()
	return size.Col(), err
}
