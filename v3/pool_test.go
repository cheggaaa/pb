//go:build linux || darwin || freebsd || netbsd || openbsd || solaris || dragonfly || plan9 || aix
// +build linux darwin freebsd netbsd openbsd solaris dragonfly plan9 aix

package pb

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

func poolTestBar(name string, clean bool) *ProgressBar {
	return ProgressBarTemplate(name).New(10).Set(CleanOnFinish, clean)
}

func TestPoolAddStartedBar(t *testing.T) {
	buf := new(bytes.Buffer)
	bar := New(10).SetWriter(buf).SetRefreshRate(time.Hour).
		Set(Terminal, true).Set(ReturnSymbol, "\r").Set(Color, false).Start()
	startTime := bar.StartTime()
	if bar.Write(); buf.Len() == 0 {
		t.Fatal("bar does not write")
	}
	buf.Reset()

	NewPool(bar)
	if bar.IsStarted() {
		t.Error("bar keeps its own writer")
	}
	if bar.IsFinished() {
		t.Error("bar is finished")
	}
	if !bar.StartTime().Equal(startTime) {
		t.Error("start time is reset")
	}
	if wipe := regexp.MustCompile(`^\r +\r$`); !wipe.Match(buf.Bytes()) {
		t.Errorf("bar's line is not wiped: %q", buf.String())
	}

	buf.Reset()
	bar.Finish()
	if buf.Len() != 0 {
		t.Errorf("bar wrote on its own: %q", buf.String())
	}
}

func TestPoolCleanOnFinish(t *testing.T) {
	root, clean, keep := poolTestBar("root", false), poolTestBar("clean", true), poolTestBar("keep", false)
	pool := NewPool(root, clean, keep)

	out, finished := pool.render(true, 0, 20)
	if finished {
		t.Error("pool is finished")
	}
	if n := strings.Count(out, "\n"); n != 3 {
		t.Errorf("first frame has %d lines: %q", n, out)
	}

	clean.Finish()
	keep.Finish()
	out, finished = pool.render(false, 0, 20)
	if finished {
		t.Error("pool is finished while root is running")
	}
	if !strings.HasPrefix(out, "\033[3A") || !strings.HasSuffix(out, "\033[J") {
		t.Errorf("frame does not rewind and erase the old one: %q", out)
	}
	if strings.Contains(out, "clean") || !strings.Contains(out, "root") || !strings.Contains(out, "keep") {
		t.Errorf("unexpected bars: %q", out)
	}
	if len(pool.bars) != 3 {
		t.Errorf("pool holds %d bars", len(pool.bars))
	}

	out, _ = pool.render(false, 0, 20)
	if !strings.HasPrefix(out, "\033[2A") || strings.Contains(out, "\033[J") {
		t.Errorf("unexpected steady frame: %q", out)
	}
}

func TestPoolAllBarsCleaned(t *testing.T) {
	bar := poolTestBar("bar", true)
	pool := NewPool(bar)
	pool.render(true, 0, 20)

	bar.Finish()
	out, finished := pool.render(false, 0, 20)
	if !finished {
		t.Error("pool is not finished")
	}
	if out != "\033[1A\r\033[J" {
		t.Errorf("unexpected frame: %q", out)
	}
	if out, _ = pool.render(false, 0, 20); out != "" {
		t.Errorf("empty pool moves the cursor: %q", out)
	}
}

func TestPoolRestartedBarReappears(t *testing.T) {
	root, reused := poolTestBar("root", false), poolTestBar("reused", true)
	pool := NewPool(root, reused)
	pool.render(true, 0, 20)

	reused.Finish()
	if out, _ := pool.render(false, 0, 20); strings.Contains(out, "reused") {
		t.Errorf("finished bar is drawn: %q", out)
	}
	reused.Start()
	if out, _ := pool.render(false, 0, 20); !strings.Contains(out, "reused") {
		t.Errorf("restarted bar is hidden: %q", out)
	}
}

func TestPoolCleanedBarsFreeRows(t *testing.T) {
	root, c1, c2 := poolTestBar("root", false), poolTestBar("c1", true), poolTestBar("c2", true)
	pool := NewPool(root, c1, c2)
	pool.render(true, 2, 20)

	c1.Finish()
	c2.Finish()
	out, finished := pool.render(false, 2, 20)
	if finished {
		t.Error("pool is finished while root is running")
	}
	if !strings.Contains(out, "root") {
		t.Errorf("root is hidden: %q", out)
	}
}

func TestPoolConcurrentAddAndFinish(t *testing.T) {
	root := poolTestBar("root", false)
	pool := NewPool(root)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bar := StartNew(10)
			bar.Set(CleanOnFinish, true)
			pool.Add(bar)
			for j := 0; j < 10; j++ {
				bar.Increment()
				time.Sleep(time.Millisecond)
			}
			bar.Finish()
		}()
	}
	wg.Wait()

	out, finished := pool.render(false, 0, 40)
	if finished {
		t.Error("pool is finished while root is running")
	}
	if !strings.Contains(out, "root") {
		t.Errorf("root is not visible: %q", out)
	}
	root.Finish()
	out, finished = pool.render(false, 0, 40)
	if !finished {
		t.Error("pool is not finished after root finish")
	}
}
