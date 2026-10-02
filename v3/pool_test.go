package pb

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPoolCleanOnFinish_SingleBar(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	bar := New(100)
	bar.Set(CleanOnFinish, true)

	pool := NewPool(bar)
	pool.Output = buf

	// First print with running bar
	finished := pool.print(true)
	if finished {
		t.Fatal("expected pool not finished")
	}
	if pool.lastBarsCount != 1 {
		t.Fatalf("expected lastBarsCount 1, got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 1 {
		t.Fatalf("expected 1 bar, got %d", len(pool.bars))
	}
	outFirst := buf.String()
	if !strings.Contains(outFirst, "0 / 100") {
		t.Fatalf("expected output to contain bar representation, got: %q", outFirst)
	}

	buf.Reset()
	// Finish the bar
	bar.Finish()

	// Next print should clear the finished bar
	finished = pool.print(false)
	if !finished {
		t.Fatal("expected pool finished")
	}
	if pool.lastBarsCount != 0 {
		t.Fatalf("expected lastBarsCount 0, got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 0 {
		t.Fatalf("expected 0 bars remaining, got %d", len(pool.bars))
	}

	outSecond := buf.String()
	// Must contain cursor up and spaces to clear line, then reposition cursor
	if !strings.Contains(outSecond, "\033[1A") {
		t.Fatalf("expected cursor move up in output, got: %q", outSecond)
	}
	if !strings.Contains(outSecond, strings.Repeat(" ", defaultBarWidth)) {
		t.Fatalf("expected cleared blank line in output, got: %q", outSecond)
	}
	if !strings.Contains(outSecond, "\033[1A\r") {
		t.Fatalf("expected cursor repositioning in output, got: %q", outSecond)
	}
}

func TestPoolCleanOnFinish_MultipleBarsIncremental(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	root := New(10)
	child1 := New(100)
	child1.Set(CleanOnFinish, true)
	child2 := New(100)
	child2.Set(CleanOnFinish, true)

	pool := NewPool(root, child1, child2)
	pool.Output = buf

	// Initial render
	finished := pool.print(true)
	if finished {
		t.Fatal("expected pool not finished")
	}
	if pool.lastBarsCount != 3 {
		t.Fatalf("expected lastBarsCount 3, got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 3 {
		t.Fatalf("expected 3 bars, got %d", len(pool.bars))
	}

	buf.Reset()
	// Finish child1
	child1.Finish()
	finished = pool.print(false)
	if finished {
		t.Fatal("expected pool not finished (root and child2 still active)")
	}
	if pool.lastBarsCount != 2 {
		t.Fatalf("expected lastBarsCount 2, got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 2 {
		t.Fatalf("expected 2 bars in pool, got %d", len(pool.bars))
	}
	out := buf.String()
	// Should move up 3 lines, draw 2 bars, clear 1 line, move up 1 line
	if !strings.HasPrefix(out, "\033[3A") {
		t.Fatalf("expected prefix \\033[3A, got: %q", out)
	}
	if !strings.Contains(out, "\033[1A\r") {
		t.Fatalf("expected cursor reposition \\033[1A\\r, got: %q", out)
	}

	buf.Reset()
	// Finish child2
	child2.Finish()
	finished = pool.print(false)
	if finished {
		t.Fatal("expected pool not finished (root still active)")
	}
	if pool.lastBarsCount != 1 {
		t.Fatalf("expected lastBarsCount 1, got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 1 {
		t.Fatalf("expected 1 bar in pool, got %d", len(pool.bars))
	}
	out = buf.String()
	// Should move up 2 lines, draw 1 bar, clear 1 line, move up 1 line
	if !strings.HasPrefix(out, "\033[2A") {
		t.Fatalf("expected prefix \\033[2A, got: %q", out)
	}
	if !strings.Contains(out, "\033[1A\r") {
		t.Fatalf("expected cursor reposition \\033[1A\\r, got: %q", out)
	}

	buf.Reset()
	// Finish root (CleanOnFinish = false)
	root.Finish()
	finished = pool.print(false)
	if !finished {
		t.Fatal("expected pool finished")
	}
	if pool.lastBarsCount != 1 {
		t.Fatalf("expected lastBarsCount 1 (root preserved), got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 1 {
		t.Fatalf("expected 1 bar remaining (root), got %d", len(pool.bars))
	}
}

func TestPoolCleanOnFinish_StopClearsBars(t *testing.T) {
	// Recreating scenario from issue #223
	buf := bytes.NewBuffer(nil)
	root := New(5)
	child1 := New(1000)
	child1.Set(CleanOnFinish, true)
	child2 := New(1000)
	child2.Set(CleanOnFinish, true)
	child3 := New(1000)
	child3.Set(CleanOnFinish, true)

	pool := NewPool(root, child1, child2, child3)
	pool.Output = buf

	// First print while all bars active
	pool.print(true)
	if pool.lastBarsCount != 4 {
		t.Fatalf("expected lastBarsCount 4, got %d", pool.lastBarsCount)
	}

	// All child bars finish
	child1.Finish()
	child2.Finish()
	child3.Finish()

	buf.Reset()
	// Final print on pool stop
	pool.print(false)
	if pool.lastBarsCount != 1 {
		t.Fatalf("expected lastBarsCount 1 (only root remains), got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 1 || pool.bars[0] != root {
		t.Fatalf("expected only root bar remaining in pool, got %v", pool.bars)
	}

	out := buf.String()
	// Cursor moves up 4 lines, draws 1 bar, clears 3 lines, moves up 3 lines
	if !strings.HasPrefix(out, "\033[4A") {
		t.Fatalf("expected prefix \\033[4A, got: %q", out)
	}
	if !strings.Contains(out, "\033[3A\r") {
		t.Fatalf("expected cursor reposition \\033[3A\\r, got: %q", out)
	}
}

func TestPoolCleanOnFinish_AllCleaned(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	b1 := New(10)
	b1.Set(CleanOnFinish, true)
	b2 := New(20)
	b2.Set(CleanOnFinish, true)

	pool := NewPool(b1, b2)
	pool.Output = buf

	pool.print(true)
	if pool.lastBarsCount != 2 {
		t.Fatalf("expected lastBarsCount 2, got %d", pool.lastBarsCount)
	}

	b1.Finish()
	b2.Finish()

	buf.Reset()
	finished := pool.print(false)
	if !finished {
		t.Fatal("expected pool finished")
	}
	if pool.lastBarsCount != 0 {
		t.Fatalf("expected lastBarsCount 0, got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 0 {
		t.Fatalf("expected 0 bars remaining, got %d", len(pool.bars))
	}
	out := buf.String()
	if !strings.HasPrefix(out, "\033[2A") {
		t.Fatalf("expected prefix \\033[2A, got: %q", out)
	}
	if !strings.Contains(out, "\033[2A\r") {
		t.Fatalf("expected cursor reposition \\033[2A\\r, got: %q", out)
	}
}

func TestPoolCleanOnFinish_WriterWorker(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	root := New(5)
	pool := NewPool(root)
	pool.Output = buf
	pool.RefreshRate = 10 * time.Millisecond
	pool.shutdownCh = make(chan struct{})
	pool.workerCh = make(chan struct{})

	go pool.writer()

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			child := New(50)
			child.Set(CleanOnFinish, true)
			pool.Add(child)
			for j := 0; j < 50; j++ {
				child.Increment()
				time.Sleep(time.Millisecond)
			}
			child.Finish()
			root.Increment()
		}()
	}

	wg.Wait()
	close(pool.shutdownCh)
	<-pool.workerCh

	pool.m.Lock()
	defer pool.m.Unlock()
	if pool.lastBarsCount != 1 {
		t.Fatalf("expected lastBarsCount 1 after stopping pool, got %d", pool.lastBarsCount)
	}
	if len(pool.bars) != 1 || pool.bars[0] != root {
		t.Fatalf("expected only root bar remaining in pool, got %v", pool.bars)
	}
}

func TestPoolCleanOnFinish_ConcurrentRace(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	root := New(10)
	pool := NewPool(root)
	pool.Output = buf
	pool.RefreshRate = 5 * time.Millisecond
	pool.shutdownCh = make(chan struct{})
	pool.workerCh = make(chan struct{})

	go pool.writer()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			child := New(20)
			child.Set(CleanOnFinish, true)
			pool.Add(child)
			for j := 0; j < 20; j++ {
				child.Increment()
				time.Sleep(time.Millisecond)
			}
			child.Finish()
		}(i)
	}

	wg.Wait()
	close(pool.shutdownCh)
	<-pool.workerCh

	pool.m.Lock()
	defer pool.m.Unlock()
	if pool.lastBarsCount != 1 {
		t.Fatalf("expected lastBarsCount 1, got %d", pool.lastBarsCount)
	}
}
