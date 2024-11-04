//go:build run
// +build run

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nyaosorg/go-readline-ny"
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
	ed.SetPredictColor(readline.PredictColorBlueItalic)

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
