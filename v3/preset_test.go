package pb

import (
	"strings"
	"testing"
)

func TestPreset(t *testing.T) {
	prefix := "Prefix"
	suffix := "Suffix"

	for _, preset := range []ProgressBarTemplate{Full, Default, Simple} {
		bar := preset.New(100).
			SetCurrent(20).
			Set("prefix", prefix).
			Set("suffix", suffix).
			SetWidth(50)

		// initialize the internal state
		_, _ = bar.render()
		s := bar.String()
		if !strings.HasPrefix(s, prefix+" ") {
			t.Error("prefix not found:", s)
		}
		if !strings.HasSuffix(s, " "+suffix) {
			t.Error("suffix not found:", s)
		}
	}
}
