package pb

import (
	"fmt"
	"time"
)

type Pool struct {
	RefreshRate time.Duration
	bars        []*ProgressBar
	isFinished  bool
}

func (p *Pool) Add(pbs ...*ProgressBar) {
	for _, bar := range pbs {
		bar.ManualUpdate = true
		bar.NotPrint = true
		bar.Start()
		p.bars = append(p.bars, bar)
	}
}

func (p *Pool) Start() (err error) {
	p.RefreshRate = DefaultRefreshRate
	quit, err := lockEcho()
	if err != nil {
		return
	}
	go p.writer(quit)
	return
}

func (p *Pool) writer(finish chan int) {
	var first = true
	var out string

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
			p.isFinished = true
			finish <- 1
			return
		}
		time.Sleep(p.RefreshRate)
	}
}
