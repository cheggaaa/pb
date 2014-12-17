package pb

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"time"
)

type Pool struct {
	RefreshRate time.Duration
	bars        []*ProgressBar
}

func (p *Pool) Add(pbs ...*ProgressBar) {
	for _, bar := range pbs {
		bar.ManualUpdate = true
		bar.NotPrint = true
		bar.Start()
		p.bars = append(p.bars, bar)
	}
}

func (p *Pool) Start() {
	p.RefreshRate = DefaultRefreshRate
	go p.writer()
}

func (p *Pool) writer() {
	var first = true
	var out string
	if oldState, err := terminal.MakeRaw(0); err != nil {
		panic(err)
	} else {
		defer terminal.Restore(0, oldState)
	}
	for {
		if first {
			first = false
		} else {
			out = fmt.Sprintf("\033[%dA", len(p.bars))
		}
		isFinished := true
		for _, bar := range p.bars {
			bar.Update()
			out += fmt.Sprintf("\r%s\n", bar.String())
			if !bar.isFinish {
				isFinished = false
			}
		}
		fmt.Print(out)
		if isFinished {
			return
		}
		time.Sleep(p.RefreshRate)
	}
}
