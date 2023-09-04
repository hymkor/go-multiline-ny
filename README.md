go-multiline-ny
===============

[![Go Reference](https://pkg.go.dev/badge/github.com/hymkor/go-multiline-ny.svg)](https://pkg.go.dev/github.com/hymkor/go-multiline-ny)

This is the readline package that supports multiline input and extends [go-readline-ny] with new keybindings. It is compatible with Emacs.

| Key | Feature
|-----|---------
| Ctrl-M or Enter | Insert a new line[^Y]
| Ctrl-J(or Ctrl-Enter[^X]) | Submit all lines
| Ctrl-P or Up | Move the cursor to the previous line
| Ctrl-N or Down | Move the cursor to the next line
| Alt-P or Ctrl-Up | Fetch the previous lines-set of the history
| Alt-N or Ctrl-Down | Fetch the next lines-set of the history
| Ctrl-Y | Paste the string in the clipboard

[go-readline-ny]: https://github.com/nyaosorg/go-readline-ny
[^X]: Only WindowsTerminal or Teraterm
[^Y]: It is possible to give the condition to submit.

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
    fmt.Println("C-m or Enter      : Insert a linefeed")
    fmt.Println("C-p or UP         : Move to the previous line.")
    fmt.Println("C-n or DOWN       : Move to the next line")
    fmt.Println("C-j               : Submit")
    fmt.Println("C-c               : Abort.")
    fmt.Println("C-D with no chars : Quit.")
    fmt.Println("C-UP   or M-P     : Move to the previous history entry")
    fmt.Println("C-DOWN or M-N     : Move to the next history entry")

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

[Terminate input only if you type Enter when it ends with a semicolon](./examples/example-swap.go)
---------

```examples/example-semi.go
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
    fmt.Println("C-m or Enter      : Submit when lines end with `;`")
    fmt.Println("                    Otherwise insert a linefeed.")
    fmt.Println("C-j               : Submit always")
    fmt.Println("C-c               : Abort.")
    fmt.Println("C-D with no chars : Quit.")

    var ed multiline.Editor
    ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
        return fmt.Fprintf(w, "[%d] ", lnum+1)
    })

    ed.SubmitOnEnterWhen(func(lines []string, _ int) bool {
        return strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), ";")
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
