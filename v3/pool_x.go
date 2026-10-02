//go:build linux || darwin || freebsd || netbsd || openbsd || solaris || dragonfly || plan9 || aix
// +build linux darwin freebsd netbsd openbsd solaris dragonfly plan9 aix

package pb

import (
	"fmt"
	"os"
	"strings"

	"github.com/cheggaaa/pb/v3/termutil"
)

func (p *Pool) print(first bool) bool {
	p.m.Lock()
	defer p.m.Unlock()
	var out string
	if !first && p.lastBarsCount > 0 {
		out = fmt.Sprintf("\033[%dA", p.lastBarsCount)
	}
	isFinished := true
	for _, bar := range p.bars {
		if !bar.IsFinished() {
			isFinished = false
			break
		}
	}
	var activeBars []*ProgressBar
	for _, bar := range p.bars {
		if !bar.IsFinished() || !bar.GetBool(CleanOnFinish) {
			activeBars = append(activeBars, bar)
		}
	}
	p.bars = activeBars
	bars := p.bars
	rows, cols, err := termutil.TerminalSize()
	if err != nil {
		cols = defaultBarWidth
	}
	if rows > 0 && len(bars) > rows {
		// we need to hide bars that overflow terminal height
		bars = bars[len(bars)-rows:]
	}
	for _, bar := range bars {
		bar.SetWidth(cols)
		result := bar.String()
		if r := cols - CellCount(result); r > 0 {
			result += strings.Repeat(" ", r)
		}
		out += fmt.Sprintf("\r%s\n", result)
	}
	if p.lastBarsCount > len(bars) {
		extraLines := p.lastBarsCount - len(bars)
		for i := 0; i < extraLines; i++ {
			out += fmt.Sprintf("\r%s\n", strings.Repeat(" ", cols))
		}
		out += fmt.Sprintf("\033[%dA\r", extraLines)
	}
	if p.Output != nil {
		fmt.Fprint(p.Output, out)
	} else {
		fmt.Fprint(os.Stderr, out)
	}
	p.lastBarsCount = len(bars)
	return isFinished
}
