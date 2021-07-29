// +build linux darwin freebsd netbsd openbsd solaris dragonfly plan9 aix

package pb

import (
	"fmt"
	"os"
	"strings"

	"github.com/cbehopkins/pb/v3/termutil"
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
	if p.width>0 {
		cols = p.width
	}
	if rows > 0 && len(bars) > rows {
		bars = bars[:rows]
	}
	for _, bar := range bars {
		if !bar.IsFinished() {
			isFinished = false
		}
		bar.SetWidth(cols)
		msg := bar.String()
		toAdd := cols - len(msg)
		if toAdd < 0 {
			toAdd = 0
		}
		out += fmt.Sprintf("\r%s%s\n", msg, strings.Repeat(" ", toAdd))
	}
	if p.Output != nil {
		fmt.Fprint(p.Output, out)
	} else {
		fmt.Fprint(os.Stderr, out)
	}
	p.lastBarsCount = len(p.bars)
	return isFinished
}
