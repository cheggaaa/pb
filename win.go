// +build windows

package pb

func bold(str string) string {
    return str
}

func terminalWidth() (int, error) {
    return 80, nil
}
