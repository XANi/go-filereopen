
# Fileopen [![GoDoc][doc-img]][doc]


Firereopen provides a write-only file interface that will reopen the given file if it is moved/removed in meantime, such as when using logrotate.

It uses inodes to see the change.

Currently it uses simple polling with settable interval.

Tested on Linux only

Example:
```go 
package main

import (
	"fmt"
	"github.com/XANi/go-filereopen"
	"time"
)

func main() {
	fd, err := filereopen.OpenFileForAppend("test.log", 0644)
	if err != nil {
		fmt.Printf("can't open: %s\n", err)
	}
	f.SetInterval(time.Second * 10) // defaults to once per second
	for {
		time.Sleep(time.Second)
		n, err := fmt.Fprintf(fd, "date: %s\n", time.Now().String())
		if err != nil {
			fmt.Printf("err: %s[%d]\n", err, n)
		}
	}
}
```

Reopen can also be triggered by `f.Reopen()`.

Underlying filedescriptor will only be swapped if opening a new one was successful. 
The old filedescriptor will be synced immediately after the swap and closed few seconds later 

Reopen is retried indefinitely, errors could be retrieved by using `SetErrorFunction` to set the error handler


[doc-img]: https://pkg.go.dev/badge/github.com/XANi/go-filereopen
[doc]: https://pkg.go.dev/github.com/XANi/go-filereopen
