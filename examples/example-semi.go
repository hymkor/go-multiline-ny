//go:build run

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-colorable"

	"github.com/nyaosorg/go-readline-ny/simplehistory"

	"github.com/hymkor/go-multiline-ny"
)

type OSClipboard struct{}

func (OSClipboard) Read() (string, error) {
	return clipboard.ReadAll()
}

func (OSClipboard) Write(s string) error {
	return clipboard.WriteAll(s)
}

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

	// Use the clipboard of the operating system.
	ed.LineEditor.Clipboard = OSClipboard{}

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
