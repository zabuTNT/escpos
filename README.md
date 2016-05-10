# escpos

##Setup

Make sure you have a working Go installation.

First of all install the dependency with
`go get github.com/tarm/serial`

Now run 
`github.com/zabuTNT/escpos`

##Example
```
package main

import (
	"fmt"
	"github.com/zabuTNT/escpos"
)

func main() {
	my_printer := escpos.NewPrinter()
	err := my_printer.SetTCPDst("192.168.1.122", "9100", 10)
	if err == nil {
		my_printer.Send("Hello World!")
		my_printer.CloseDst()
	} else {
		fmt.Println(err)
	}

}
```
