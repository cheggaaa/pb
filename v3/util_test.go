package pb

import (
	"github.com/fatih/color"
	"testing"
)

var testColorString = color.RedString("red") +
	color.GreenString("hello") +
	"simple" +
	color.WhiteString("進捗")

func TestUtilCellCount(t *testing.T) {
	if e, l := 18, CellCount(testColorString); l != e {
		t.Errorf("Invalid length %d, expected %d", l, e)
	}
}

func TestUtilStripString(t *testing.T) {
	if r, e := StripString("12345", 4), "1234"; r != e {
		t.Errorf("Invalid result '%s', expected '%s'", r, e)
	}

	if r, e := StripString("12345", 5), "12345"; r != e {
		t.Errorf("Invalid result '%s', expected '%s'", r, e)
	}
	if r, e := StripString("12345", 10), "12345"; r != e {
		t.Errorf("Invalid result '%s', expected '%s'", r, e)
	}

	s := color.RedString("1") + "23"
	e := color.RedString("1") + "2"
	if r := StripString(s, 2); r != e {
		t.Errorf("Invalid result '%s', expected '%s'", r, e)
	}
	return
}

func TestUtilRound(t *testing.T) {
	if v := round(4.4); v != 4 {
		t.Errorf("Unexpected result: %v", v)
	}
	if v := round(4.501); v != 5 {
		t.Errorf("Unexpected result: %v", v)
	}
}

func TestUtilFormatBytes(t *testing.T) {
	inputs := []struct {
		v int64
		s bool
		e string
	}{
		{v: 1000, s: false, e: "1000 B"},
		{v: 1024, s: false, e: "1.00 KiB"},
		{v: 3*_MiB + 140*_KiB, s: false, e: "3.14 MiB"},
		{v: 2 * _GiB, s: false, e: "2.00 GiB"},
		{v: 2048 * _GiB, s: false, e: "2.00 TiB"},

		{v: 999, s: true, e: "999 B"},
		{v: 1024, s: true, e: "1.02 kB"},
		{v: 3*_MB + 140*_kB, s: true, e: "3.14 MB"},
		{v: 2 * _GB, s: true, e: "2.00 GB"},
		{v: 2048 * _GB, s: true, e: "2.05 TB"},
	}

	for _, input := range inputs {
		actual := formatBytes(input.v, input.s)
		if actual != input.e {
			t.Errorf("Expected {%s} was {%s}", input.e, actual)
		}
	}
}

func BenchmarkUtilsCellCount(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		CellCount(testColorString)
	}
}
