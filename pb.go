package pb

import (
	"time"
	"sync/atomic"
	"runtime"
	"syscall"
	"unsafe"
	"strings"
	"fmt"
)

var (
	// Default refresh rate - 200ms
	DefaultRefreshRate = time.Millisecond * 200
	
	BarStart = " ["
	BarEnd   = "] "
	Empty    = "_"
	Current  = "="
	CurrentN = ">"
)

const (	
	TIOCGWINSZ = 0x5413
	TIOCGWINSZ_OSX = 1074295912
)

// Create new progress bar object
func New(total int) (*ProgressBar) {
	return &ProgressBar{
		Total : int64(total),
		RefreshRate : DefaultRefreshRate,
		ShowPercent : true,
		ShowCounters: true,
		ShowBar     : true,
	}
}

// Create new object and start 
func StartNew(total int) (pb *ProgressBar) {
	pb = New(total)
	pb.Start()
	return
}

type ProgressBar struct {
	Total int64
	RefreshRate time.Duration
	ShowPercent, ShowCounters, ShowBar bool
	current int64	
	isFinish bool
}

// Start print
func (pb *ProgressBar) Start() {
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
	fmt.Println()
}

// End print and write string 'str'
func (pb *ProgressBar) FinishPrint(str string) {
	pb.Finish()
	fmt.Println(bold(str))
}


func (pb *ProgressBar) write(current int64) {
	width, _ := terminalWidth()
	var percentBox, countersBox, barBox, end, out string
	
	// percents
	if pb.ShowPercent {
		percent := float64(current) / (float64(pb.Total) / float64(100))
		percentBox = fmt.Sprintf("%#.02f %% ", percent)
	}
	
	// counters
	if pb.ShowCounters {
		countersBox = bold(fmt.Sprintf("%d / %d", current, pb.Total))
	}
	
	// bar
	if pb.ShowBar {
		size := width - len(countersBox + BarStart + BarEnd + percentBox)
		if size > 0 {
			curCount := int(float64(current) / (float64(pb.Total) / float64(size)))
    		emptCount := size - curCount
    		barBox = BarStart
    		if emptCount == 0 {
    				barBox += strings.Repeat(Current, curCount)
    		} else if curCount > 0 {
    				barBox += strings.Repeat(Current, curCount - 1) + CurrentN
    		}
    		barBox += strings.Repeat(Empty, emptCount) + BarEnd
		}
	} 
	
	// check len
	out = countersBox + barBox + percentBox
	if len(out) < width {
		end = strings.Repeat(" ", width - len(out))
	}
	
	// bold
	if countersBox != "" {
		countersBox = bold(countersBox)
	}
	if percentBox != "" {
		percentBox = bold(percentBox)
	}
	out = countersBox + barBox + percentBox
	
	// and print!
	fmt.Print("\r" + out + end)
}

func (pb *ProgressBar) writer() {
	var c, oc int64
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

func bold(str string) string {
    return "\033[1m" + str + "\033[0m"
}

type window struct {
    Row    uint16
    Col    uint16
    Xpixel uint16
    Ypixel uint16
}

func terminalWidth() (int, error) {
    w := new(window)
    tio := TIOCGWINSZ
    if runtime.GOOS == "darwin" {
    	tio = TIOCGWINSZ_OSX
    }
    res, _, err := syscall.Syscall(syscall.SYS_IOCTL,
        uintptr(syscall.Stdin),
        uintptr(tio),
        uintptr(unsafe.Pointer(w)),
    )
    if int(res) == -1 {
        return 0, err
    }
    return int(w.Col), nil
}