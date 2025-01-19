go-multiline-ny
===============

[![Go Reference](https://pkg.go.dev/badge/github.com/hymkor/go-multiline-ny.svg)](https://pkg.go.dev/github.com/hymkor/go-multiline-ny)

This is the readline package that supports multiline input and extends [go-readline-ny] with new keybindings. It is compatible with Emacs.

| Key | Feature
|-----|---------
| Ctrl-M or Enter | Insert a new line[^Y]
| Ctrl-J/Enter or Escape-Enter | Submit all lines
| Ctrl-P or Up   | Move cursor to previous line or last line of previous set of inputs in history
| Ctrl-N or Down | Move cursor to next line or first line of next set of inputs in history
| Alt-P or Ctrl-Up | Fetch the previous lines-set of the history
| Alt-N or Ctrl-Down | Fetch the next lines-set of the history
| Ctrl-Y | Paste the string in the clipboard
| Ctrl-R | Incremental search

[go-readline-ny]: https://github.com/nyaosorg/go-readline-ny
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
    "regexp"
    "strings"

    "github.com/mattn/go-colorable"

    "github.com/nyaosorg/go-readline-ny"
    "github.com/nyaosorg/go-readline-ny/simplehistory"

    "github.com/hymkor/go-multiline-ny"
)

func main() {
    ctx := context.Background()
    fmt.Println("C-m or Enter      : Insert a linefeed")
    fmt.Println("C-p or UP         : Move to the previous line.")
    fmt.Println("C-n or DOWN       : Move to the next line")
    fmt.Println("C-j or Esc+Enter  : Submit")
    fmt.Println("C-c               : Abort.")
    fmt.Println("C-D with no chars : Quit.")
    fmt.Println("C-UP   or M-P     : Move to the previous history entry")
    fmt.Println("C-DOWN or M-N     : Move to the next history entry")

    var ed multiline.Editor
    ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
        return fmt.Fprintf(w, "[%d] ", lnum+1)
    })
    ed.SetPredictColor(readline.PredictColorBlueItalic)

    ed.LineEditor.Highlight = []readline.Highlight{
        // Double quotation -> Magenta
        {Pattern: regexp.MustCompile(`"([^"]*\\")*[^"]*$|"([^"]*\\")*[^"]*"`), Sequence: "\x1B[35;49;1m"},
        // Enviroment variable -> Cyan
        {Pattern: regexp.MustCompile(`%[^%]*$|%[^%]*%`), Sequence: "\x1B[36;49;1m"},
    }
    ed.LineEditor.ResetColor = "\x1B[0m"
    ed.LineEditor.DefaultColor = "\x1B[37;49;1m"

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

Acknowledgements
----------------

- [@spiegel-im-spiegel](https://github.com/spiegel-im-spiegel)
- [@apstndb](https://github.com/apstndb)

Release notes
-------------

- [Release notes (English)](./release_note_en.md)
- [Release notes (Japanese)](./release_note_ja.md)
