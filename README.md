go-multiline-ny
===============

[![Go Reference](https://pkg.go.dev/badge/github.com/hymkor/go-multiline-ny.svg)](https://pkg.go.dev/github.com/hymkor/go-multiline-ny)

This is the readline package that supports multiline input and extends [go-readline-ny] with new keybindings. It is compatible with Emacs.

| Key | Feature
|-----|---------
| Ctrl-M or Enter | Insert a new line
| Ctrl-J(or Ctrl-Enter[^X]) | Submit all lines
| Ctrl-P or Up | Move the cursor to the previous line
| Ctrl-N or Down | Move the cursor to the next line
| Alt-P or Ctrl-Up | Fetch the previous lines-set of the history
| Alt-N or Ctrl-Down | Fetch the next lines-set of the history
| Ctrl-Y | Paste the string in the clipboard

[go-readline-ny]: https://github.com/nyaosorg/go-readline-ny
[^X]: Only WindowsTerminal or Teraterm

[Example](./examples/example.go)
---------

![image](./demo.gif)

```examples/example.go
package main

import (
    "context"
    "fmt"
    "io"
    "os"
    "strings"

    "github.com/hymkor/go-multiline-ny"
    "github.com/mattn/go-colorable"
    "github.com/nyaosorg/go-readline-ny/simplehistory"
)

func main() {
    ctx := context.Background()
    fmt.Println("Ctrl-M or Enter      : Insert a linefeed")
    fmt.Println("Ctrl-P or UP         : Move to the previous line.")
    fmt.Println("Ctrl-N or DOWN       : Move to the next line")
    fmt.Println("Ctrl-J               : Submit")
    fmt.Println("Ctrl-C               : Abort.")
    fmt.Println("Ctrl-D with no chars : Quit.")
    fmt.Println("Ctrl-UP   or ALT-P   : Move to the previous history entry")
    fmt.Println("Ctrl-DOWN or ALT-N   : Move to the next history entry")

    var ed multiline.Editor
    ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
        return fmt.Fprintf(w, "[%d] ", lnum+1)
    })

    // To enable escape sequence on Windows.
    // (On other operating systems, it can be ommited)
    ed.SetWriter(colorable.NewColorableStdout())

    history := simplehistory.New()
    ed.SetHistory(history)
    ed.SetHistoryCycling(true)

    for {
        lines, err := ed.Read(ctx)
        if err != nil {
            fmt.Fprintln(os.Stderr, err.Error())
            return
        }
        L := strings.Join(lines, "\n")
        fmt.Println("-----")
        fmt.Println(L)
        fmt.Println("-----")
        history.Add(L)
    }
}
```

[Example to swap Ctrl-J and Ctrl-M](./examples/example-swap.go)
---------

```examples/example-swap.go
package main

import (
    "context"
    "fmt"
    "os"
    "strings"

    "github.com/hymkor/go-multiline-ny"
    "github.com/mattn/go-colorable"
    "github.com/nyaosorg/go-readline-ny"
    "github.com/nyaosorg/go-readline-ny/keys"
    "github.com/nyaosorg/go-readline-ny/simplehistory"
)

func main() {
    var ed multiline.Editor

    ed.SetWriter(colorable.NewColorableStdout())
    history := simplehistory.New()
    ed.SetHistory(history)
    ed.SetHistoryCycling(true)

    ed.BindKey(keys.CtrlM, readline.AnonymousCommand(ed.Submit))
    ed.BindKey(keys.CtrlJ, readline.AnonymousCommand(ed.NewLine))

    ctx := context.Background()
    for {
        lines, err := ed.Read(ctx)
        if err != nil {
            fmt.Fprintln(os.Stderr, err.Error())
            return
        }
        L := strings.Join(lines, "\n")
        fmt.Println("-----")
        fmt.Println(L)
        fmt.Println("-----")
        history.Add(L)
    }
}
```
