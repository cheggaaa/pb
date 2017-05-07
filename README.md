# Terminal progress bar for Go  

### Unstable! Under development!

This is proposal for the second version of progress bar   
- based on text/template   
- can take custom elements   
- using colors is easy   

## Installation

```
go get gopkg.in/cheggaaa/pb.v2
```   

## Usage   

```Go
package main

import (
	"gopkg.in/cheggaaa/pb.v2"
	"time"
)

func main() {
	progress := new(pb.ProgressBar)
	tmpl := `{{ string . "prefix"}}{{counters . | red}} {{ bar . "" "" (cycle . "↖" "↗" "↘" "↙" )}} {{percent .}}`
	var n = 1000
	progress.SetTotal(int64(n)).SetTemplate(tmpl).Start()
	for i := 0; i < n; i++ {
		progress.Increment()
		time.Sleep(time.Millisecond * 20)
		switch i {
		case 100:
			progress.Set("prefix", "i > 100 ")
		case 500:
			progress.Set("prefix", "i > 500 ")
		}
	}
	progress.Finish()
}

```