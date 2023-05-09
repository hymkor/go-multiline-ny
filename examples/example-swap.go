//go:build run
// +build run

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
