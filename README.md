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
    fmt.Println("Enter, DOWN or Ctrl-N: New line or move to the next line")
    fmt.Println("UP or Ctrl-P: Move to the previous line.")
    fmt.Println("Ctrl-Enter: Sumbit")
    fmt.Println("Ctrl-C: Cancel lines.")
    fmt.Println("Ctrl-D: Quit.")
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