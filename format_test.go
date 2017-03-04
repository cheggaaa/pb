package pb_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

func Test_DefaultsToInteger(t *testing.T) {
	value := int64(1000)
	expected := strconv.Itoa(int(value))
	actual := pb.Format(value).String()

	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_CanFormatAsInteger(t *testing.T) {
	value := int64(1000)
	expected := strconv.Itoa(int(value))
	actual := pb.Format(value).To(pb.U_NO).String()

	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_CanFormatAsBytes(t *testing.T) {
	inputs := []struct {
		v int64
		e string
	}{
		{v: 1000, e: "1000 B"},
		{v: 1024, e: "1.00 KiB"},
		{v: 3*pb.MiB + 140*pb.KiB, e: "3.14 MiB"},
		{v: 2 * pb.GiB, e: "2.00 GiB"},
		{v: 2048 * pb.GiB, e: "2.00 TiB"},
	}

	for _, input := range inputs {
		actual := pb.Format(input.v).To(pb.U_BYTES).String()
		if actual != input.e {
			t.Error(fmt.Sprintf("Expected {%s} was {%s}", input.e, actual))
		}
	}
}

func Test_CanFormatDuration(t *testing.T) {
	value := 10 * time.Minute
	expected := "10m0s"
	actual := pb.Format(int64(value)).To(pb.U_DURATION).String()
	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_DefaultUnitsWidth(t *testing.T) {
	value := 10
	expected := "     10"
	actual := pb.Format(int64(value)).Width(7).String()
	if actual != expected {
		t.Error(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}
