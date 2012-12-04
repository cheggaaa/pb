## Terminal progress bar for Go  

### Installation
```
go get github.com/cheggaaa/pb
```   

### Usage   
```Go
package main

import (
	"github.com/cheggaaa/pb"
	"time"
)

func main() {
	count := 100000
	bar := pb.StartNew(count)
	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond)
	}
	bar.FinishPrint("The End!")
}
```   
Result will be like this:
```
> go run test.go
23976 / 100000 [==============>___________________________________________________] 23.98 %
```
