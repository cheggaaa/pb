// +build windows
// +build !appengine

package pb

import (
	"os"

	"github.com/olekukonko/ts"
)

var tty = os.Stdin

// terminalWidth returns width of the terminal.
func terminalWidth() (int, error) {
	size, err := ts.GetSize()
	return size.Col(), err
}
