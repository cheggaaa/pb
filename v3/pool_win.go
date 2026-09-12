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
	if !first {
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
	visible := 0
	for _, bar := range p.bars {
		if !bar.IsFinished() {
			isFinished = false
		}
		// honor CleanOnFinish: drop finished bars from the pool frame
		if bar.IsFinished() && bar.GetBool(CleanOnFinish) {
			continue
		}
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
	if p.Output != nil {
		fmt.Fprint(p.Output, out)
	} else {
		fmt.Print(out)
	}
	if cleared := p.lastBarsCount - visible; cleared > 0 {
		coords, err := termutil.GetCursorPos()
		if err == nil {
			coords.Y -= int16(cleared)
			if coords.Y < 0 {
				coords.Y = 0
			}
			coords.X = 0
			_ = termutil.SetCursorPos(coords)
		}
	}
	p.lastBarsCount = visible
	return isFinished
}
