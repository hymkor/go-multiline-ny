//go:build run

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hymkor/go-multiline-ny"
	"github.com/mattn/go-colorable"
)

func main() {
	var ed multiline.Editor

	ed.SetWriter(colorable.NewColorableStdout())

	ctx := context.Background()
	lines := []string{ "Default1","Default2","Default3"}
	for{
		var err error
		lines, err = ed.Read(ctx,lines...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		L := strings.Join(lines, "\n")
		fmt.Println("-----")
		fmt.Println(L)
		fmt.Println("-----")

	}
}

