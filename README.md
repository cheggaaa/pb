# Terminal progress bar for Go  
[![Coverage Status](https://coveralls.io/repos/github/cheggaaa/pb/badge.svg?branch=v2)](https://coveralls.io/github/cheggaaa/pb?branch=v2)

## Installation

```
go get github.com/cheggaaa/pb/v3
```   

Documentation for v1 bar available [here](README_V1.md)

## Quick start   

```Go
package main

import (
	"time"
	
	"github.com/cheggaaa/pb/v3"
)

func main() {
	count := 100000
	// create and start new bar
	bar := pb.StartNew(count)
	
	// start bar from 'default' template
	// bar := pb.Default.Start(count)
	
	// start bar from 'simple' template
	// bar := pb.Simple.Start(count)
	
	// start bar from 'full' template
	// bar := pb.Full.Start(count)
	
	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond)
	}
	bar.Finish()
}

```

Result will be like this:

```
> go run test.go
37158 / 100000 [================>_______________________________] 37.16% 1m11s
```

## Settings

```Go  
// create bar
bar := pb.New(count)

// refresh info every second (default 200ms)
bar.SetRefreshRate(time.Second)

// force set io.Writer, by default it's os.Stderr
bar.SetWriter(os.Stdout)

// bar will format numbers as bytes (B, Kb, Mb, etc)
bar.Set(pb.Byte, true)

// set custom bar template
bar.SetTemplateString(myTemplate)

// check for error after template set
if err = bar.Err(); err != nil {
    return
}

// start bar
bar.Start()

``` 

## Progress bar for IO Operations
```go
package main

import (
	"crypto/rand"
	"io"
	"io/ioutil"

	"github.com/cheggaaa/pb/v3"
)

func main() {

	var limit int64 = 1024 * 1024 * 500
	// we will copy 200 Mb from /dev/rand to /dev/null
	reader := io.LimitReader(rand.Reader, limit)
	writer := ioutil.Discard

	// start new bar
	bar := pb.Full.Start64(limit)
	// create proxy reader
	barReader := bar.NewProxyReader(reader)
	// copy from proxy reader
	io.Copy(writer, barReader)
	// finish bar
	bar.Finish()
}

```

## Custom Progress Bar templates

Rendering based on builtin text/template package. You can use existing pb's elements or create you own.

All available elements are described in element.go file.  

#### All in one example:
```go
tmpl := `{{ red "With funcs:" }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{speed . | rndcolor }} {{percent .}} {{string . "my_green_string" | green}} {{string . "my_blue_string" | blue}}`
// start bar based on our template
bar := pb.ProgressBarTemplate(tmpl).Start64(limit)
// set values for string elements
bar.Set("my_green_string", "green").
	Set("my_blue_string", "blue")
```