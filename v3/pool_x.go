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
	if !first {
		out = fmt.Sprintf("\033[%dA", p.lastBarsCount)
	}
	isFinished := true
	bars := p.bars
	rows, cols, err := termutil.TerminalSize()
	if err != nil {
		cols = defaultBarWidth
	}
	if rows > 0 && len(bars) > rows {
		// we need to hide bars that overflow terminal height
		bars = bars[len(bars)-rows:]
	}
	visible := 0
	for _, bar := range bars {
		if !bar.IsFinished() {
			isFinished = false
		}
		// honor CleanOnFinish: drop finished bars from the pool frame
		if bar.IsFinished() && bar.GetBool(CleanOnFinish) {
			continue
		}
		bar.SetWidth(cols)
		result := bar.String()
		if r := cols - CellCount(result); r > 0 {
			result += strings.Repeat(" ", r)
		}
		out += fmt.Sprintf("\r%s\n", result)
		visible++
	}
	// wipe lines left over when the visible count shrinks
	for i := visible; i < p.lastBarsCount; i++ {
		out += "\r" + strings.Repeat(" ", cols) + "\n"
	}
	if cleared := p.lastBarsCount - visible; cleared > 0 {
		out += fmt.Sprintf("\033[%dA", cleared)
	}
	if p.Output != nil {
		fmt.Fprint(p.Output, out)
	} else {
		fmt.Fprint(os.Stderr, out)
	}
	p.lastBarsCount = visible
	return isFinished
}
