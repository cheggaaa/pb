package pb

import (
	"bytes"
	"strings"
	"testing"
)

// TestPoolBasic - Repeat beasic test from TestPBBasic but using a pool
func TestPoolBasic(t *testing.T) {
	bar := new(ProgressBar)
	pool := NewPool(bar)
	pool.Start()
	var a, e int64
	if a, e = bar.Total(), 0; a != e {
		t.Errorf("Unexpected total: actual: %v; expected: %v", a, e)
	}
	if a, e = bar.Current(), 0; a != e {
		t.Errorf("Unexpected current: actual: %v; expected: %v", a, e)
	}
	bar.SetCurrent(10).SetTotal(20)
	if a, e = bar.Total(), 20; a != e {
		t.Errorf("Unexpected total: actual: %v; expected: %v", a, e)
	}
	if a, e = bar.Current(), 10; a != e {
		t.Errorf("Unexpected current: actual: %v; expected: %v", a, e)
	}
	bar.Add(5)
	if a, e = bar.Current(), 15; a != e {
		t.Errorf("Unexpected current: actual: %v; expected: %v", a, e)
	}
	bar.Increment()
	if a, e = bar.Current(), 16; a != e {
		t.Errorf("Unexpected current: actual: %v; expected: %v", a, e)
	}
}

type testStruct struct {
	required bool
	txt      string
}

func expectedStringsCheck(b bytes.Buffer, expected []testStruct, t *testing.T) {
	for _, ts := range expected {
		if !(strings.Contains(b.String(), ts.txt) == ts.required) {
			errTxt := " contains "
			if ts.required {
				errTxt = " does not contain "
			}
			t.Error(b.String(), errTxt, ts.txt)
		}
	}
}
func TestPoolDelete(t *testing.T) {
	bars := make([]*ProgressBar, 3)
	for i := range bars {
		bars[i] = new(ProgressBar)
		// Needed because default width is different to the width in the dummy terminal

	}

	pool := NewPool(bars...)
	pool.SetWidth(80)
	var b bytes.Buffer
	pool.Output = &b
	// We should:
	//  pool.Start()
	// But this needs access to the raw terminal, so, we do the init
	// and worker manually
	pool.workerCh = make(chan struct{})

	defer pool.Stop()
	defer func() {
		close(pool.workerCh)
	}()

	bars[0].SetCurrent(50).SetTotal(100)
	bars[1].SetCurrent(60).SetTotal(200)
	bars[2].SetCurrent(20).SetTotal(200)

	expected := []testStruct{
		{true, "50 / 100 [--------------------------->____________________________] 50.00% ? p/s"},
		{true, "60 / 200 [---------------->_______________________________________] 30.00% ? p/s"},
		{true, "20 / 200 [----->__________________________________________________] 10.00% ? p/s"},
	}

	// Check the initial setup is as expected
	if len(pool.bars) != len(bars) {
		t.Error("Bar length problem start")
	}
	pool.print(true)
	expectedStringsCheck(b, expected, t)

	// Remove a bar and make sure it goes
	err := pool.Remove(bars[1])
	if err != nil {
		t.Error("pool failed to rmove:", err)
	}
	expected[1].required = false
	if len(pool.bars) != (len(bars) - 1) {
		t.Error("Bar length problem end")
	}

	b.Reset()
	pool.print(true)
	expectedStringsCheck(b, expected, t)

	// Add a nw bar and test that works as expected
	newBar := new(ProgressBar)
	pool.Add(newBar)
	newBar.SetCurrent(50).SetTotal(1000)
	expected = append(expected, testStruct{true, "50 / 1000 [-->_____________________________________________________] 5.00% ? p/s"})

	b.Reset()
	pool.print(true)
	expectedStringsCheck(b, expected, t)

	/////////////////////////////////////////////////////////////
	// All bars that we have ever used, be sure to shut them down
	for i := range bars {
		bars[i].Finish()
	}
}

func TestPoolVariety(t *testing.T) {
	bars := make([]*ProgressBar, 5)
	for i := range bars {
		bars[i] = new(ProgressBar)
		// Needed because default width is different to the width in the dummy terminal
	}

	pool := NewPool(bars...)
	pool.SetWidth(80)
	var b bytes.Buffer
	pool.Output = &b
	// We should:
	// pool.Start()
	// But this needs access to the raw terminal, so, we do the init
	// and worker manually
	pool.workerCh = make(chan struct{})

	defer pool.Stop()
	defer func() {
		close(pool.workerCh)
	}()

	bars[0].SetCurrent(50).SetTotal(100)
	bars[1].SetCurrent(60).SetTotal(200)
	bars[1].Set("prefix", "bob:")
	bars[2].SetCurrent(6).SetTotal(10)
	bars[2].Set(Bytes, true)
	bars[3].SetCurrent(1 << 12).SetTotal(1 << 14)
	bars[3].Set(Bytes, true)
	bars[4].SetTemplateString(`{{string . "my custom text"}}`)
	bars[4].Set("my custom text", "Initialzing discombobulator")

	expected := []testStruct{
		{true, "50 / 100 [--------------------------->____________________________] 50.00% ? p/s"},
		{true, "bob:60 / 200 [--------------->____________________________________] 30.00% ? p/s"},
		{true, "6 B / 10 B [-------------------------------->_____________________] 60.00% ? p/s"},
		{true, "4.00 KiB / 16.00 KiB [---------->_________________________________] 25.00% ? p/s"},
		{true, "Initialzing discombobulator"},
	}

	pool.print(true)
	expectedStringsCheck(b, expected, t)

	/////////////////////////////////////////////////////////////
	// All bars that we have ever used, be sure to shut them down
	for i := range bars {
		bars[i].Finish()
	}
}
