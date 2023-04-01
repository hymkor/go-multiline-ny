//go:build ignore

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hymkor/go-multiline-ny"
	"github.com/mattn/go-colorable"
)

func main() {
	ctx := context.Background()
	fmt.Println("Ctrl-M or Enter      : Insert a linefeed")
	fmt.Println("Ctrl-N or DOWN       : Move to the next line")
	fmt.Println("Ctrl-P or UP         : Move to the previous line.")
	fmt.Println("Ctrl-J or Ctrl-Enter : Submit")
	fmt.Println("Ctrl-C               : Cancel lines.")
	fmt.Println("Ctrl-D with no chars : Quit.")

	var editor multiline.Editor
	editor.Prompt = func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "[%d] ", lnum+1)
	}

	// To enable escape sequence on Windows.
	// (On other operating systems, it can be ommited)
	editor.LineEditor.Writer = colorable.NewColorableStdout()

	for {
		lines, err := editor.Read(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		fmt.Println("-----")
		fmt.Println(strings.Join(lines, "\n"))
		fmt.Println("-----")
	}
}
