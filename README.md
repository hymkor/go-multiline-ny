go-multiline-ny
===============

This is the readline package supporting multi-lines that extends [go-readline-ny]

[go-readline-ny]: https://github.com/nyaosorg/go-readline-ny

![image](./demo.gif)

```example.go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/hymkor/go-multiline-ny"
)

func main() {
    ctx := context.Background()
    fmt.Println("Ctrl-M or Enter      : Insert a linefeed")
    fmt.Println("Ctrl-N or DOWN       : Move to the next line")
    fmt.Println("Ctrl-P or UP         : Move to the previous line.")
    fmt.Println("Ctrl-J or Ctrl-Enter : Submit")
    fmt.Println("Ctrl-C               : Cancel lines.")
    fmt.Println("Ctrl-D with no chars : Quit.")
    for {
        lines, err := multiline.Read(ctx)
        if err != nil {
            fmt.Fprintln(os.Stderr, err.Error())
            return
        }
        fmt.Println("-----")
        for len(lines) > 0 && lines[len(lines)-1] == "" {
            lines = lines[:len(lines)-1]
        }
        for _, s := range lines {
            fmt.Println(s)
        }
        fmt.Println("-----")
    }
}
```
