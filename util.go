package pb

import (
	"fmt"
	"gopkg.in/mattn/go-runewidth.v0"
	"math"
	"regexp"
)

const (
	_KiB = 1024
	_MiB = 1048576
	_GiB = 1073741824
	_TiB = 1099511627776
)

var ctrlFinder = regexp.MustCompile("\x1b\x5b[0-9]+\x6d")

func cellCount(s string) int {
	return runewidth.StringWidth(s)
}

func cellCountStripASCIISeq(s string) int {
	n := cellCount(s)
	for _, sm := range ctrlFinder.FindAllString(s, -1) {
		n -= cellCount(sm)
	}
	return n
}

func round(val float64) (newVal float64) {
	roundOn := 0.5
	places := 0
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

// Convert bytes to human readable string. Like a 2 MiB, 64.2 KiB, 52 B
func formatBytes(i int64) (result string) {
	switch {
	case i >= _TiB:
		result = fmt.Sprintf("%.02f TiB", float64(i)/_TiB)
	case i >= _GiB:
		result = fmt.Sprintf("%.02f GiB", float64(i)/_GiB)
	case i >= _MiB:
		result = fmt.Sprintf("%.02f MiB", float64(i)/_MiB)
	case i >= _KiB:
		result = fmt.Sprintf("%.02f KiB", float64(i)/_KiB)
	default:
		result = fmt.Sprintf("%d B", i)
	}
	return
}
