//go:build run
// +build run

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

	var editor multiline.Editor
	editor.SetPrompt(func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "[%d] ", lnum+1)
	})

	// To enable escape sequence on Windows.
	// (On other operating systems, it can be ommited)
	editor.SetWriter(colorable.NewColorableStdout())

	history := simplehistory.New()
	editor.SetHistory(history)
	editor.SetHistoryCycling(true)

	for {
		lines, err := editor.Read(ctx)
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
