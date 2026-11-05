//go:build linux || darwin || freebsd || netbsd || openbsd || solaris || dragonfly || plan9 || aix
// +build linux darwin freebsd netbsd openbsd solaris dragonfly plan9 aix

package pb

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
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
		t.Fatal("bar doesn't write")
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
		t.Errorf("frame doesn't rewind and erase the old one: %q", out)
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

// lateWriter counts writes that land after done is set
type lateWriter struct{ done, late atomic.Int32 }

func (w *lateWriter) Write(p []byte) (int, error) {
	if w.done.Load() == 1 {
		w.late.Add(1)
	}
	return len(p), nil
}

func TestPoolAddRacesStartAndFinish(t *testing.T) {
	for i := 0; i < 2000; i++ {
		w := new(lateWriter)
		bar := New(10).SetWriter(w).SetRefreshRate(10*time.Microsecond).
			Set(Terminal, true).Set(ReturnSymbol, "\r").Set(Color, false).Start()
		pool := new(Pool)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); pool.Add(bar) }()
		go func() {
			defer wg.Done()
			if i%2 == 0 {
				bar.Start()
				return
			}
			bar.Finish()
			w.done.Store(1)
		}()
		wg.Wait()
		if err := bar.Err(); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if bar.IsStarted() {
			t.Fatalf("iteration %d: bar keeps its own writer", i)
		}
		if n := w.late.Load(); n > 0 {
			t.Fatalf("iteration %d: %d writes after Finish returned", i, n)
		}
	}
}
