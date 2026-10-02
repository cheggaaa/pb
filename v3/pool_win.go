//go:build windows
// +build windows

package pb

import (
	"fmt"
	"log"
	"strings"

	"github.com/cheggaaa/pb/v3/termutil"
)

func (p *Pool) print(first bool) bool {
	p.m.Lock()
	defer p.m.Unlock()
	var out string
	if !first && p.lastBarsCount > 0 {
		coords, err := termutil.GetCursorPos()
		if err != nil {
			log.Panic(err)
		}
		coords.Y -= int16(p.lastBarsCount)
		if coords.Y < 0 {
			coords.Y = 0
		}
		coords.X = 0

		err = termutil.SetCursorPos(coords)
		if err != nil {
			log.Panic(err)
		}
	}
	cols, err := termutil.TerminalWidth()
	if err != nil {
		cols = defaultBarWidth
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

	for _, bar := range bars {
		bar.SetWidth(cols)
		result := bar.String()
		if r := cols - CellCount(result); r > 0 {
			result += strings.Repeat(" ", r)
		}
		out += fmt.Sprintf("\r%s\n", result)
	}
	extraLines := p.lastBarsCount - len(bars)
	if extraLines > 0 {
		for i := 0; i < extraLines; i++ {
			out += fmt.Sprintf("\r%s\n", strings.Repeat(" ", cols))
		}
	}
	if p.Output != nil {
		fmt.Fprint(p.Output, out)
	} else {
		fmt.Print(out)
	}
	if extraLines > 0 {
		if coords, err := termutil.GetCursorPos(); err == nil {
			coords.Y -= int16(extraLines)
			if coords.Y < 0 {
				coords.Y = 0
			}
			coords.X = 0
			_ = termutil.SetCursorPos(coords)
		}
	}
	p.lastBarsCount = len(bars)
	return isFinished
}
