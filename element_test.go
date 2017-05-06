package pb

import (
	"fmt"
	"testing"
	"time"
)

func testState(total, value int64, maxWidth int, bools ...bool) (s *State) {
	s = &State{
		total:           total,
		current:         value,
		adaptiveElWidth: maxWidth,
		ProgressBar:     new(ProgressBar),
	}
	if len(bools) > 0 {
		s.Set(Bytes, bools[0])
	}
	if len(bools) > 1 && bools[1] {
		s.adaptive = true
	}
	return
}

func testElementBarString(t *testing.T, state *State, el Element, want string, args ...string) {
	if state.ProgressBar == nil {
		state.ProgressBar = new(ProgressBar)
	}
	res := el.ProgressElement(state, args...)
	if res != want {
		t.Errorf("Unexpected result: '%s'; want: '%s'", res, want)
	}
	if state.IsAdaptiveWidth() && state.AdaptiveElWidth() != state.CellCount(res) {
		t.Errorf("Unepected width: %d; want: %d", state.CellCount(res), state.AdaptiveElWidth())
	}
}

func TestElementPercent(t *testing.T) {
	testElementBarString(t, testState(100, 50, 0), ElementPercent, "50.00%")
	testElementBarString(t, testState(100, 50, 0), ElementPercent, "50 percent", "%v percent")
	testElementBarString(t, testState(0, 50, 0), ElementPercent, "?%")
	testElementBarString(t, testState(0, 50, 0), ElementPercent, "unkn", "%v%%", "unkn")
}

func TestElementCounters(t *testing.T) {
	testElementBarString(t, testState(100, 50, 0), ElementCounters, "50 / 100")
	testElementBarString(t, testState(100, 50, 0), ElementCounters, "50 of 100", "%s of %s")
	testElementBarString(t, testState(100, 50, 0, true), ElementCounters, "50 B of 100 B", "%s of %s")
	testElementBarString(t, testState(100, 50, 0, true), ElementCounters, "50 B / 100 B")
	testElementBarString(t, testState(0, 50, 0, true), ElementCounters, "50 B")
	testElementBarString(t, testState(0, 50, 0, true), ElementCounters, "50 B / ?", "", "%[1]s / ?")
}

func TestElementBar(t *testing.T) {
	// short
	testElementBarString(t, testState(100, 50, 1, false, true), ElementBar, "[")
	testElementBarString(t, testState(100, 50, 2, false, true), ElementBar, "[]")
	testElementBarString(t, testState(100, 50, 3, false, true), ElementBar, "[>]")
	testElementBarString(t, testState(100, 50, 4, false, true), ElementBar, "[>_]")
	testElementBarString(t, testState(100, 50, 5, false, true), ElementBar, "[->_]")
	// middle
	testElementBarString(t, testState(100, 50, 10, false, true), ElementBar, "[--->____]")
	testElementBarString(t, testState(100, 50, 10, false, true), ElementBar, "<--->____>", "<", "", "", "", ">")
	// empty
	testElementBarString(t, testState(0, 50, 10, false, true), ElementBar, "[________]")
	// full
	testElementBarString(t, testState(20, 20, 10, false, true), ElementBar, "[------->]")
	// everflow
	testElementBarString(t, testState(20, 50, 10, false, true), ElementBar, "[------->]")
	// small width
	testElementBarString(t, testState(20, 50, 2, false, true), ElementBar, "[]")
	testElementBarString(t, testState(20, 50, 1, false, true), ElementBar, "[")
	// negative counters
	testElementBarString(t, testState(-50, -150, 10, false, true), ElementBar, "[------->]")
	testElementBarString(t, testState(-150, -50, 10, false, true), ElementBar, "[-->_____]")
	testElementBarString(t, testState(50, -150, 10, false, true), ElementBar, "[------->]")
	testElementBarString(t, testState(-50, 150, 10, false, true), ElementBar, "[------->]")
	// long entities / unicode
	f1 := []string{"進捗|", "многобайт", "active", "пусто", "|end"}
	testElementBarString(t, testState(100, 50, 1, false, true), ElementBar, " ", f1...)
	testElementBarString(t, testState(100, 50, 3, false, true), ElementBar, "進 ", f1...)
	testElementBarString(t, testState(100, 50, 4, false, true), ElementBar, "進捗", f1...)
	testElementBarString(t, testState(100, 50, 29, false, true), ElementBar, "進捗|многactiveпустопусто|end", f1...)
	testElementBarString(t, testState(100, 50, 11, false, true), ElementBar, "進捗|aп|end", f1...)

	// unicode
	f2 := []string{"⚑", "⚒", "⚟", "⟞", "⚐"}
	testElementBarString(t, testState(100, 50, 10, false, true), ElementBar, "⚑⚒⚒⚒⚟⟞⟞⟞⟞⚐", f2...)

	// no adaptive
	testElementBarString(t, testState(0, 50, 10), ElementBar, "[____________________________]")

	var formats = [][]string{
		[]string{},
		f1, f2,
	}

	// all widths / extreme values
	// check for panic and correct width
	for _, f := range formats {
		for tt := int64(-2); tt < 12; tt++ {
			for v := int64(-2); v < 12; v++ {
				state := testState(tt, v, 0, false, true)
				for w := -2; w < 20; w++ {
					state.adaptiveElWidth = w
					res := ElementBar(state, f...)
					var we = w
					if we <= 0 {
						we = 30
					}
					if state.CellCount(res) != we {
						t.Errorf("Unexpected len(%d): '%s'", we, res)
					}
				}
			}
		}
	}
}

func TestElementSpeed(t *testing.T) {
	var state = testState(1000, 0, 0, false)
	state.ProgressBar.startTime = time.Now()
	for i := int64(0); i < 10; i++ {
		state.current += 42
		r := ElementSpeed(state)
		if i <= 1 {
			// do not calc first two results
			if w := "? p/s"; r != w {
				t.Errorf("Unexpected result[%d]: '%s' vs '%s'", i, r, w)
			}
		} else {
			if w := "42 p/s"; r != w {
				t.Errorf("Unexpected result[%d]: '%s' vs '%s'", i, r, w)
			}
		}
		var add = -time.Second
		if i > 7 {
			add = add / 2
		}
		getSpeedObj(state).prevTime = time.Now().Add(add)
	}
}

func TestElementRemainingTime(t *testing.T) {
	var state = testState(1000, 0, 0, false)
	state.ProgressBar.startTime = time.Now()
	for i := int64(0); i < 10; i++ {
		state.current += 42
		r := ElementRemainingTime(state)
		if i <= 1 {
			// do not calc first two results
			if w := "?"; r != w {
				t.Errorf("Unexpected result[%d]: '%s' vs '%s'", i, r, w)
			}
		} else {
			w := fmt.Sprintf("%ds", 22-i)
			if r != w {
				t.Errorf("Unexpected result[%d]: '%s' vs '%s'", i, r, w)
			}
		}
		getSpeedObj(state).prevTime = time.Now().Add(-time.Second)
	}
}

func TestElementElapsedTime(t *testing.T) {
	var state = testState(1000, 0, 0, false)
	state.ProgressBar.startTime = time.Now().Truncate(time.Second)
	for i := int64(0); i < 10; i++ {
		r := ElementElapsedTime(state)
		if w := fmt.Sprintf("%ds", i); r != w {
			t.Errorf("Unexpected result[%d]: '%s' vs '%s'", i, r, w)
		}
		state.ProgressBar.startTime = state.ProgressBar.startTime.Add(-time.Second)
	}
}

func TestElementString(t *testing.T) {
	var state = testState(0, 0, 0, false)
	testElementBarString(t, state, ElementString, "", "myKey")
	state.Set("myKey", "my value")
	testElementBarString(t, state, ElementString, "my value", "myKey")
	state.Set("myKey", "my value1")
	testElementBarString(t, state, ElementString, "my value1", "myKey")
	testElementBarString(t, state, ElementString, "")
}

func TestElementCycle(t *testing.T) {
	var state = testState(0, 0, 0, false)
	testElementBarString(t, state, ElementCycle, "")
	testElementBarString(t, state, ElementCycle, "1", "1", "2", "3")
	testElementBarString(t, state, ElementCycle, "2", "1", "2", "3")
	testElementBarString(t, state, ElementCycle, "3", "1", "2", "3")
	testElementBarString(t, state, ElementCycle, "1", "1", "2", "3")
	testElementBarString(t, state, ElementCycle, "2", "1", "2")
	testElementBarString(t, state, ElementCycle, "1", "1", "2")
}

func TestAdaptiveWrap(t *testing.T) {
	var state = testState(0, 0, 0, false)
	state.first = true
	state.Set("myKey", "my value")
	el := adaptiveWrap(ElementString)
	testElementBarString(t, state, el, adElPlaceholder, "myKey")
	if v := state.recalc[0].ProgressElement(state); v != "my value" {
		t.Errorf("Unexpected result: %s", v)
	}
	state.first = false
	testElementBarString(t, state, el, adElPlaceholder, "myKey1")
	state.Set("myKey", "my value1")
	if v := state.recalc[0].ProgressElement(state); v != "my value1" {
		t.Errorf("Unexpected result: %s", v)
	}
}

func BenchmarkBar(b *testing.B) {
	var formats = map[string][]string{
		"simple":      []string{".", ".", ".", ".", "."},
		"unicode":     []string{"⚑", "⚒", "⚟", "⟞", "⚐"},
		"long":        []string{"..", "..", "..", "..", ".."},
		"longunicode": []string{"⚑⚑", "⚒⚒", "⚟⚟", "⟞⟞", "⚐⚐"},
	}
	for _, asciiSeq := range []bool{false, true} {
		for name, args := range formats {
			state := testState(100, 50, 100, false, true)
			if asciiSeq {
				state.Set(Terminal, true)
				name += "/terminal"
			}
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					ElementBar(state, args...)
				}
			})
		}
	}
}
