package pb

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"gopkg.in/cheggaaa/pb.v2/termutil"
	"gopkg.in/mattn/go-colorable.v0"
	"gopkg.in/mattn/go-isatty.v0"
)

// Version of ProgressBar library
const Version = "2.0.1"

type key int

const (
	// Bytes means we're working with byte sizes. Numbers will print as Kb, Mb, etc
	// bar.Set(pb.Bytes, true)
	Bytes key = 1 << iota

	// Terminal means we're will print to terminal and can use ascii sequences
	// Also we're will try to use terminal width
	Terminal

	// Static means progress bar will not update automaticly
	Static

	// ReturnSymbol - by default in terminal mode it's '\r'
	ReturnSymbol

	// Color by default is true when output is tty, but you can set to false for disabling colors
	Color
)

const (
	defaultBarWidth    = 100
	defaultRefreshRate = time.Millisecond * 200
)

// ProgressBar is the main object of bar
type ProgressBar struct {
	current, total int64
	width          int
	mu             sync.RWMutex
	vars           map[interface{}]interface{}
	elements       map[string]Element
	output         io.Writer
	coutput        io.Writer
	nocoutput      io.Writer
	startTime      time.Time
	refreshRate    time.Duration
	tmpl           *template.Template
	state          *State
	buf            *bytes.Buffer
	ticker         *time.Ticker
	finish         chan struct{}
	finished       bool
	configured     bool
	err            error
}

func (pb *ProgressBar) configure() {
	if pb.configured {
		return
	}
	pb.configured = true

	if pb.vars == nil {
		pb.vars = make(map[interface{}]interface{})
	}
	if pb.output == nil {
		pb.output = os.Stderr
	}

	if pb.tmpl == nil {
		var err error
		pb.tmpl, err = getTemplate(Default, nil)
		if err != nil {
			panic(err)
		}
	}
	if pb.vars[Terminal] == nil {
		if f, ok := pb.output.(*os.File); ok {
			if isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd()) {
				pb.vars[Terminal] = true
			}
		}
	}
	if pb.vars[ReturnSymbol] == nil {
		if tm, ok := pb.vars[Terminal].(bool); ok && tm {
			pb.vars[ReturnSymbol] = "\r"
		}
	}
	if pb.vars[Color] == nil {
		if tm, ok := pb.vars[Terminal].(bool); ok && tm {
			pb.vars[Color] = true
		}
	}
	if pb.refreshRate == 0 {
		pb.refreshRate = defaultRefreshRate
	}
	if f, ok := pb.output.(*os.File); ok {
		pb.coutput = colorable.NewColorable(f)
	} else {
		pb.coutput = pb.output
	}
	pb.nocoutput = colorable.NewNonColorable(pb.output)
}

// Start starts the bar
func (pb *ProgressBar) Start() *ProgressBar {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.configure()
	pb.finished = false
	pb.state = nil
	if st, ok := pb.vars[Static].(bool); ok && st {
		return pb
	}
	pb.finish = make(chan struct{})
	pb.ticker = time.NewTicker(pb.refreshRate)
	go pb.writer(pb.finish)
	return pb
}

func (pb *ProgressBar) writer(finish chan struct{}) {
	for {
		select {
		case <-pb.ticker.C:
			pb.write()
		case <-finish:
			pb.ticker.Stop()
			pb.write()
			finish <- struct{}{}
			return
		}
	}
}

func (pb *ProgressBar) write() {
	result := pb.render()
	if pb.Err() != nil {
		return
	}
	if ret, ok := pb.Get(ReturnSymbol).(string); ok {
		result += ret
		if ret == "\r" {
			pb.mu.RLock()
			if pb.finished {
				result += "\n"
			}
			pb.mu.RUnlock()
		}
	}
	if pb.GetBool(Color) {
		pb.coutput.Write([]byte(result))
	} else {
		pb.nocoutput.Write([]byte(result))
	}
}

// Total return current total bar value
func (pb *ProgressBar) Total() int64 {
	return atomic.LoadInt64(&pb.total)
}

// SetTotal sets the total bar value
func (pb *ProgressBar) SetTotal(value int64) *ProgressBar {
	atomic.StoreInt64(&pb.total, value)
	return pb
}

func (pb *ProgressBar) SetCurrent(value int64) *ProgressBar {
	atomic.StoreInt64(&pb.current, value)
	return pb
}

// Add adding given int64 value to bar value
func (pb *ProgressBar) Add64(value int64) *ProgressBar {
	atomic.AddInt64(&pb.current, value)
	return pb
}

// Add adding given int value to bar value
func (pb *ProgressBar) Add(value int) *ProgressBar {
	return pb.Add64(int64(value))
}

func (pb *ProgressBar) Increment() *ProgressBar {
	return pb.Add64(1)
}

// Set sets any value by any key
func (pb *ProgressBar) Set(key, value interface{}) *ProgressBar {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if pb.vars == nil {
		pb.vars = make(map[interface{}]interface{})
	}
	pb.vars[key] = value
	return pb
}

// Get return value by key
func (pb *ProgressBar) Get(key interface{}) interface{} {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	if pb.vars == nil {
		return nil
	}
	return pb.vars[key]
}

// GetBool return value by key and try to convert there to boolean
// If value doesn't set or not boolean - return false
func (pb *ProgressBar) GetBool(key interface{}) bool {
	if v, ok := pb.Get(key).(bool); ok {
		return v
	}
	return false
}

// SetWidth sets the bar width
// When given value <= 0 would be using the terminal width (if possible) or default value.
func (pb *ProgressBar) SetWidth(width int) *ProgressBar {
	pb.mu.Lock()
	pb.width = width
	pb.mu.Unlock()
	return pb
}

// Width return the bar width
// It's current terminal width or settled over 'SetWidth' value.
func (pb *ProgressBar) Width() (width int) {
	defer func() {
		if r := recover(); r != nil {
			width = defaultBarWidth
		}
	}()
	pb.mu.RLock()
	width = pb.width
	pb.mu.RUnlock()
	if width <= 0 {
		var err error
		if width, err = termutil.TerminalWidth(); err != nil {
			return defaultBarWidth
		}
	}
	return
}

// SetWriter sets the io.Writer. Bar will write in this writer
// By default this is os.Stderr
func (pb *ProgressBar) SetWriter(w io.Writer) *ProgressBar {
	pb.mu.Lock()
	pb.output = w
	pb.configure()
	pb.mu.Unlock()
	return pb
}

// StartTime return the time when bar started
func (pb *ProgressBar) StartTime() time.Time {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return pb.startTime
}

// Format convert int64 to string according to the current settings
func (pb *ProgressBar) Format(v int64) string {
	if pb.GetBool(Bytes) {
		return formatBytes(v)
	}
	return strconv.FormatInt(v, 10)
}

// Finish stops the bar
func (pb *ProgressBar) Finish() *ProgressBar {
	pb.mu.Lock()
	if pb.finished {
		pb.mu.Unlock()
		return pb
	}
	pb.finished = true
	finishChan := pb.finish
	pb.mu.Unlock()
	if finishChan != nil {
		finishChan <- struct{}{}
		<-finishChan
	}
	return pb
}

// CellCount calculates string width in cells
func (pb *ProgressBar) CellCount(s string) int {
	if pb.GetBool(Terminal) {
		return cellCountStripASCIISeq(s)
	}
	return cellCount(s)
}

// SetTemplate sets ProgressBar tempate string and parse it
func (pb *ProgressBar) SetTemplate(tmpl string) *ProgressBar {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.tmpl, pb.err = getTemplate(tmpl, pb.elements)
	return pb
}

func (pb *ProgressBar) render() (result string) {
	pb.mu.Lock()
	pb.configure()
	if pb.state == nil {
		pb.state = &State{first: true, ProgressBar: pb}
		pb.buf = bytes.NewBuffer(nil)
	} else {
		pb.state.first = false
	}
	pb.state.finished = pb.finished
	pb.mu.Unlock()

	pb.state.width = pb.Width()
	pb.state.total = atomic.LoadInt64(&pb.total)
	pb.state.current = atomic.LoadInt64(&pb.current)
	pb.buf.Reset()

	if e := pb.tmpl.Execute(pb.buf, pb.state); e != nil {
		pb.SetErr(e)
		return ""
	}

	result = pb.buf.String()

	aec := len(pb.state.recalc)
	if aec == 0 {
		// no adaptive elements
		// just return result
		return
	}

	staticWidth := pb.CellCount(result) - (aec * adElPlaceholderLen)
	pb.state.adaptiveElWidth = (pb.state.width - staticWidth) / aec
	for _, el := range pb.state.recalc {
		result = strings.Replace(result, adElPlaceholder, el.ProgressElement(pb.state), 1)
	}
	pb.state.recalc = pb.state.recalc[:0]
	return
}

func (pb *ProgressBar) SetErr(err error) *ProgressBar {
	pb.mu.Lock()
	pb.err = err
	pb.mu.Unlock()
	return pb
}

// Err return possible error
// When all ok - will be nil
// May contain template.Execute errors
func (pb *ProgressBar) Err() error {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return pb.err
}

// String return currrent string representation of ProgressBar
func (pb *ProgressBar) String() string {
	return pb.render()
}

// ProgressElement implements Element interface
func (pb *ProgressBar) ProgressElement(s *State, args ...string) string {
	if s.IsAdaptiveWidth() {
		pb.SetWidth(s.AdaptiveElWidth())
	}
	return pb.String()
}

// State represents the current state of bar
// Need for bar elements
type State struct {
	*ProgressBar

	total, current int64

	width, adaptiveElWidth int

	first, finished, adaptive bool

	recalc []Element
}

// Total it's bar int64 total
func (s *State) Total() int64 {
	return s.total
}

// Value it's current value
func (s *State) Value() int64 {
	return s.current
}

// Width of bar
func (s *State) Width() int {
	return s.width
}

// AdaptiveElWidth - adaptive elements must return string with given cell count (when AdaptiveElWidth > 0)
func (s *State) AdaptiveElWidth() int {
	return s.adaptiveElWidth
}

// IsAdaptiveWidth returns true when element must be shown as adaptive
func (s *State) IsAdaptiveWidth() bool {
	return s.adaptive
}

// IsFinished return true when bar is finished
func (s *State) IsFinished() bool {
	return s.finished
}

// IsFirst return true only in first render
func (s *State) IsFirst() bool {
	return s.first
}
