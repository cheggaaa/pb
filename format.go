package pb

import (
	"fmt"
	"strings"
	"time"
)

type Units int

const (
	// By default, without type handle
	U_NO Units = iota
	// Handle as b, Kb, Mb, etc
	U_BYTES
)

// Format integer
func Format(i int64, units Units, width int) string {
	switch units {
	case U_BYTES:
		return FormatBytes(i)
	default:
		// by default just convert to string
		return fmt.Sprintf(fmt.Sprintf("%%%dd", width), i)
	}
}

// Convert bytes to human readable string. Like a 2 MB, 64.2 KB, 52 B
func FormatBytes(i int64) (result string) {
	switch {
	case i > (1024 * 1024 * 1024 * 1024):
		result = fmt.Sprintf("%.02f TB", float64(i)/1024/1024/1024/1024)
	case i > (1024 * 1024 * 1024):
		result = fmt.Sprintf("%.02f GB", float64(i)/1024/1024/1024)
	case i > (1024 * 1024):
		result = fmt.Sprintf("%.02f MB", float64(i)/1024/1024)
	case i > 1024:
		result = fmt.Sprintf("%.02f KB", float64(i)/1024)
	default:
		result = fmt.Sprintf("%d B", i)
	}
	result = strings.Trim(result, " ")
	return
}

func FormatDuration(d time.Duration) string {
	res := ""
	if d > time.Hour*24 {
		res = fmt.Sprintf("%dd", d/24/time.Hour)
		d -= (d / time.Hour / 24) * (time.Hour * 24)
	}
	return fmt.Sprintf("%s%v ", res, d)
}
