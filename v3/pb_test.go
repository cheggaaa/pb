package pb

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
)

func TestPBBasic(t *testing.T) {
	bar := new(ProgressBar)
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

func TestPBWidth(t *testing.T) {
	terminalWidth = func() (int, error) {
		return 50, nil
	}
	// terminal width
	bar := new(ProgressBar)
	if a, e := bar.Width(), 50; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
	// terminal width error
	terminalWidth = func() (int, error) {
		return 0, errors.New("test error")
	}
	if a, e := bar.Width(), defaultBarWidth; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
	// terminal width panic
	terminalWidth = func() (int, error) {
		panic("test")
		return 0, nil
	}
	if a, e := bar.Width(), defaultBarWidth; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
	// set negative terminal width
	bar.SetWidth(-42)
	if a, e := bar.Width(), defaultBarWidth; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
	// set terminal width
	bar.SetWidth(42)
	if a, e := bar.Width(), 42; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
}

func TestPBMaxWidth(t *testing.T) {
	terminalWidth = func() (int, error) {
		return 50, nil
	}
	// terminal width
	bar := new(ProgressBar)
	if a, e := bar.Width(), 50; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}

	bar.SetMaxWidth(55)
	if a, e := bar.Width(), 50; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}

	bar.SetMaxWidth(38)
	if a, e := bar.Width(), 38; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
}

func TestAddTotal(t *testing.T) {
	bar := new(ProgressBar)
	bar.SetTotal(0)
	bar.AddTotal(50)
	got := bar.Total()
	if got != 50 {
		t.Errorf("bar.Total() = %v, want %v", got, 50)
	}
	bar.AddTotal(-10)
	got = bar.Total()
	if got != 40 {
		t.Errorf("bar.Total() = %v, want %v", got, 40)
	}
}

func TestPBTemplate(t *testing.T) {
	defer setUnicodeProgressBarEnv("false")()

	bar := new(ProgressBar)
	result := bar.SetTotal(100).SetCurrent(50).SetWidth(40).String()
	expected := "50 / 100 [------->________] 50.00% ? p/s"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}

	// check strip
	result = bar.SetWidth(8).String()
	expected = "50 / 100"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}

	// invalid template
	for _, invalidTemplate := range []string{
		`{{invalid template`, `{{speed}}`,
	} {
		bar.SetTemplateString(invalidTemplate)
		result = bar.String()
		expected = ""
		if result != expected {
			t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
		}
		if err := bar.Err(); err == nil {
			t.Errorf("Must be error")
		}
	}

	// simple template without adaptive elemnts
	bar.SetTemplateString(`{{counters . }}`)
	result = bar.String()
	expected = "50 / 100"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}
}

func TestUnicodeProgressBarEnvUsesFiraDefaultBarElements(t *testing.T) {
	defer setUnicodeProgressBarEnv("true")()

	for name, tmpl := range map[string]ProgressBarTemplate{
		"Full":    Full,
		"Default": Default,
		"Simple":  Simple,
		"Custom":  `{{bar . }}`,
	} {
		result := tmpl.New(100).SetCurrent(0).SetWidth(60).String()
		if !strings.Contains(result, "") {
			t.Errorf("%s must use fira empty left border: %q", name, result)
		}
		if !strings.Contains(result, "") {
			t.Errorf("%s must use fira right border: %q", name, result)
		}
	}
}

func TestUnicodeProgressBarEnvIgnoresOne(t *testing.T) {
	defer setUnicodeProgressBarEnv("1")()

	result := ProgressBarTemplate(`{{bar . }}`).New(100).SetCurrent(0).SetWidth(10).String()
	if result != "[________]" {
		t.Errorf("UNICODE_PROGRESS_BAR=1 must keep ascii defaults: %q", result)
	}
}

func TestUnicodeProgressBarEnvDoesNotOverrideExplicitBarArgs(t *testing.T) {
	defer setUnicodeProgressBarEnv("true")()

	for _, test := range []struct {
		name     string
		template ProgressBarTemplate
		current  int64
		finished bool
		expected string
	}{
		{
			name:     "five args empty",
			template: `{{bar . "<" "=" ">" "." ">"}}`,
			current:  0,
			expected: "<........>",
		},
		{
			name:     "five args finished",
			template: `{{bar . "<" "=" ">" "." ">"}}`,
			current:  100,
			finished: true,
			expected: "<========>",
		},
		{
			name:     "explicit empty extras",
			template: `{{bar . "<" "=" ">" "." ">" "" ""}}`,
			current:  0,
			expected: "<........>",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			bar := New64(100).SetTemplate(test.template).SetCurrent(test.current).SetWidth(10)
			if test.finished {
				bar.Finish()
			}

			result := bar.String()
			if result != test.expected {
				t.Errorf("explicit bar args must be used: %q; want %q", result, test.expected)
			}
		})
	}
}

func TestPBStartFinish(t *testing.T) {
	bar := ProgressBarTemplate(`{{counters . }}`).New(0)
	for i := int64(0); i < 2; i++ {
		if bar.IsStarted() {
			t.Error("Must be false")
		}
		var buf = bytes.NewBuffer(nil)
		bar.SetTotal(100).
			SetCurrent(int64(i)).
			SetWidth(7).
			Set(Terminal, true).
			SetWriter(buf).
			SetRefreshRate(time.Millisecond * 20).
			Start()
		if !bar.IsStarted() {
			t.Error("Must be true")
		}
		time.Sleep(time.Millisecond * 100)
		bar.Finish()
		if buf.Len() == 0 {
			t.Error("no writes")
		}
		var resultsString = strings.TrimPrefix(buf.String(), "\r")
		if !strings.HasSuffix(resultsString, "\n") {
			t.Error("No end \\n symb")
		} else {
			resultsString = resultsString[:len(resultsString)-1]
		}
		var results = strings.Split(resultsString, "\r")
		if len(results) < 3 {
			t.Errorf("Unexpected writes count: %v", len(results))
		}
		exp := fmt.Sprintf("%d / 100", i)
		for i, res := range results {
			if res != exp {
				t.Errorf("Unexpected result[%d]: '%v'", i, res)
			}
		}
		// test second finish call
		bar.Finish()
	}
}

func TestPBFlags(t *testing.T) {
	// Static
	color.NoColor = false
	buf := bytes.NewBuffer(nil)
	bar := ProgressBarTemplate(`{{counters . | red}}`).New(100)
	bar.Set(Static, true).SetCurrent(50).SetWidth(10).SetWriter(buf).Start()
	if bar.IsStarted() {
		t.Error("Must be false")
	}
	bar.Write()
	result := buf.String()
	expected := "50 / 100"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n'%s'\n'%s'", result, expected)
	}
	if !bar.state.IsFirst() {
		t.Error("must be true")
	}
	// Color
	bar.Set(Color, true)
	buf.Reset()
	bar.Write()
	result = buf.String()
	expected = color.RedString("50 / 100")
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n'%s'\n'%s'", result, expected)
	}
	if bar.state.IsFirst() {
		t.Error("must be false")
	}
	// Terminal
	bar.Set(Terminal, true).SetWriter(buf)
	buf.Reset()
	bar.Write()
	result = buf.String()
	expected = "\r" + color.RedString("50 / 100") + "  "
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n'%s'\n'%s'", result, expected)
	}
}

func setEnv(key, value string) func() {
	old, ok := os.LookupEnv(key)
	os.Setenv(key, value)

	return func() {
		if ok {
			os.Setenv(key, old)
			return
		}
		os.Unsetenv(key)
	}
}

func setUnicodeProgressBarEnv(value string) func() {
	restoreEnv := setEnv(unicodeProgressBarEnv, value)
	oldDefaultBarEls := defaultBarEls
	configureDefaultBarEls()

	return func() {
		defaultBarEls = oldDefaultBarEls
		restoreEnv()
	}
}

func BenchmarkRender(b *testing.B) {
	var formats = []string{
		string(Simple),
		string(Default),
		string(Full),
		`{{string . "prefix" | red}}{{counters . | green}} {{bar . | yellow}} {{percent . | cyan}} {{speed . | cyan}}{{string . "suffix" | cyan}}`,
	}
	var names = []string{
		"Simple", "Default", "Full", "Color",
	}
	for i, tmpl := range formats {
		bar := new(ProgressBar)
		bar.SetTemplateString(tmpl).SetWidth(100)
		b.Run(names[i], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bar.String()
			}
		})
	}
}
