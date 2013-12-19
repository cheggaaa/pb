package pb

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

var (
	// Default refresh rate - 200ms
	DefaultRefreshRate = time.Millisecond * 200

	BarStart = "["
	BarEnd   = "]"
	Empty    = "_"
	Current  = "="
	CurrentN = ">"
)

// Create new progress bar object
func New(total int) *ProgressBar {
	return &ProgressBar{
		Total:        int64(total),
		RefreshRate:  DefaultRefreshRate,
		ShowPercent:  true,
		ShowCounters: true,
		ShowBar:      true,
		ShowTimeLeft: true,
	}
}

// Create new object and start
func StartNew(total int) (pb *ProgressBar) {
	pb = New(total)
	pb.Start()
	return
}

// Callback for custom output
// For example:
// bar.Callback = func(s string) {
//     mySuperPrint(s)
// }
//
type Callback func(out string)

type ProgressBar struct {
	current int64 // current must be first member of struct (https://code.google.com/p/go/issues/detail?id=5278)

	Total                            int64
	RefreshRate                      time.Duration
	ShowPercent, ShowCounters        bool
	ShowSpeed, ShowTimeLeft, ShowBar bool
	Output                           io.Writer
	Callback                         Callback
	NotPrint                         bool
	Units                            int

	isFinish  bool
	startTime time.Time
}

// Start print
func (pb *ProgressBar) Start() {
	pb.startTime = time.Now()
	if pb.Total == 0 {
		pb.ShowBar = false
		pb.ShowTimeLeft = false
		pb.ShowPercent = false
	}
	go pb.writer()
}

// Increment current value
func (pb *ProgressBar) Increment() int {
	return pb.Add(1)
}

// Set current value
func (pb *ProgressBar) Set(current int) {
	atomic.StoreInt64(&pb.current, int64(current))
}

// Add to current value
func (pb *ProgressBar) Add(add int) int {
	return int(atomic.AddInt64(&pb.current, int64(add)))
}

// End print
func (pb *ProgressBar) Finish() {
	pb.isFinish = true
	pb.write(atomic.LoadInt64(&pb.current))
	if !pb.NotPrint {
		fmt.Println()
	}
}

// End print and write string 'str'
func (pb *ProgressBar) FinishPrint(str string) {
	pb.Finish()
	fmt.Println(str)
}

// implement io.Writer
func (pb *ProgressBar) Write(p []byte) (n int, err error) {
	n = len(p)
	pb.Add(n)
	return
}

// implement io.Reader
func (pb *ProgressBar) Read(p []byte) (n int, err error) {
	n = len(p)
	pb.Add(n)
	return
}

func (pb *ProgressBar) write(current int64) {
	width, _ := terminalWidth()
	var percentBox, countersBox, timeLeftBox, speedBox, barBox, end, out string

	// percents
	if pb.ShowPercent {
		percent := float64(current) / (float64(pb.Total) / float64(100))
		percentBox = fmt.Sprintf(" %#.02f %% ", percent)
	}

	// counters
	if pb.ShowCounters {
		if pb.Total > 0 {
			countersBox = fmt.Sprintf("%s / %s ", Format(current, pb.Units), Format(pb.Total, pb.Units))
		} else {
			countersBox = Format(current, pb.Units) + " "
		}
	}

	// time left
	if pb.ShowTimeLeft && current > 0 {
		fromStart := time.Now().Sub(pb.startTime)
		perEntry := fromStart / time.Duration(current)
		left := time.Duration(pb.Total-current) * perEntry
		left = (left / time.Second) * time.Second
		if left > 0 {
			timeLeftBox = left.String()
		}
	}

	// speed
	if pb.ShowSpeed && current > 0 {
		fromStart := time.Now().Sub(pb.startTime)
		speed := float64(current) / (float64(fromStart) / float64(time.Second))
		speedBox = Format(int64(speed), pb.Units) + "/s "
	}

	// bar
	if pb.ShowBar {
		size := width - len(countersBox+BarStart+BarEnd+percentBox+timeLeftBox+speedBox)
		if size > 0 {
			curCount := int(math.Ceil((float64(current) / float64(pb.Total)) * float64(size)))
			emptCount := size - curCount
			barBox = BarStart
			if emptCount < 0 {
				emptCount = 0
			}
			if curCount > size {
				curCount = size
			}
			if emptCount <= 0 {
				barBox += strings.Repeat(Current, curCount)
			} else if curCount > 0 {
				barBox += strings.Repeat(Current, curCount-1) + CurrentN
			}

			barBox += strings.Repeat(Empty, emptCount) + BarEnd
		}
	}

	// check len
	out = countersBox + barBox + percentBox + speedBox + timeLeftBox
	if len(out) < width {
		end = strings.Repeat(" ", width-len(out))
	}

	out = countersBox + barBox + percentBox + speedBox + timeLeftBox

	// and print!
	switch {
	case pb.Output != nil:
		fmt.Fprint(pb.Output, out+end)
	case pb.Callback != nil:
		pb.Callback(out + end)
	case !pb.NotPrint:
		fmt.Print("\r" + out + end)
	}
}

func (pb *ProgressBar) writer() {
	var c, oc int64
	oc = -1
	for {
		if pb.isFinish {
			break
		}
		c = atomic.LoadInt64(&pb.current)
		if c != oc {
			pb.write(c)
			oc = c
		}
		time.Sleep(pb.RefreshRate)
	}
}

type window struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}
