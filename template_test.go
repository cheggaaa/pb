package pb

import (
	"testing"
)

func TestProgressBarTemplate(t *testing.T) {
	// test New
	bar := ProgressBarTemplate(`{{counters . }}`).New()
	result := bar.String()
	expected := "0"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}
	if bar.IsStarted() {
		t.Error("Must be false")
	}

	// test Start
	bar = ProgressBarTemplate(`{{counters . }}`).Start(42)
	result = bar.String()
	expected = "0 / 42"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}
	if !bar.IsStarted() {
		t.Error("Must be true")
	}
}
