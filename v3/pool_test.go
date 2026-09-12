package pb

import (
	"bytes"
	"strings"
	"testing"
)

func TestPoolCleanOnFinish(t *testing.T) {
	var buf bytes.Buffer
	root := New(1)
	clean := New(10)
	clean.Set(CleanOnFinish, true)
	keep := New(10)

	pool := &Pool{Output: &buf}
	pool.Add(root, clean, keep)

	root.SetCurrent(1)
	clean.SetCurrent(10).Finish()
	keep.SetCurrent(5)

	pool.print(true)
	out := buf.String()
	if strings.Contains(out, "10 / 10") {
		t.Fatalf("finished CleanOnFinish bar still rendered:\n%s", out)
	}
	if !strings.Contains(out, "5 / 10") {
		t.Fatalf("active bar missing from pool output:\n%s", out)
	}
	if pool.lastBarsCount != 2 {
		t.Fatalf("lastBarsCount=%d, want 2 (root + keep)", pool.lastBarsCount)
	}

	// bar without CleanOnFinish stays visible after Finish
	keep.SetCurrent(10).Finish()
	buf.Reset()
	pool.print(true)
	out = buf.String()
	if !strings.Contains(out, "10 / 10") {
		t.Fatalf("finished bar without CleanOnFinish should remain:\n%s", out)
	}
	if pool.lastBarsCount != 2 {
		t.Fatalf("lastBarsCount=%d, want 2", pool.lastBarsCount)
	}
}

func TestPoolCleanOnFinishClearsShrinkingFrame(t *testing.T) {
	var buf bytes.Buffer
	root := New(2)
	a := New(5)
	a.Set(CleanOnFinish, true)
	b := New(5)
	b.Set(CleanOnFinish, true)

	pool := &Pool{Output: &buf}
	pool.Add(root, a, b)

	pool.print(true)
	if pool.lastBarsCount != 3 {
		t.Fatalf("lastBarsCount=%d, want 3", pool.lastBarsCount)
	}

	a.SetCurrent(5).Finish()
	b.SetCurrent(5).Finish()
	root.SetCurrent(2)

	buf.Reset()
	pool.print(false)
	out := buf.String()
	if pool.lastBarsCount != 1 {
		t.Fatalf("lastBarsCount=%d after clean, want 1", pool.lastBarsCount)
	}
	if strings.Contains(out, "5 / 5") {
		t.Fatalf("cleaned bars still present:\n%s", out)
	}
	// unix path blanks leftover lines then moves cursor back up
	if !strings.Contains(out, "\033[2A") {
		t.Fatalf("expected cursor adjust after wiping 2 leftover lines:\n%q", out)
	}
}
