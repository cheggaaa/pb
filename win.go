// +build windows

package pb

func bold(str string) string {
    return str
}

func terminalWidth() (int, error) {
    80, nil
}
