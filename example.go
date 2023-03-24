//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nyaosorg/go-readline-ny"

	"github.com/hymkor/go-multiline-ny"
)

func mains() error {
	editor := &readline.Editor{}
	ctx := context.Background()
	fmt.Println("Enter, DOWN or Ctrl-N: New line or move to the next line")
	fmt.Println("UP or Ctrl-P: Move to the previous line.")
	fmt.Println("Ctrl-Enter: Sumbit")
	fmt.Println("Ctrl-C: Cancel lines.")
	fmt.Println("Ctrl-D: Quit.")
	for {
		lines, err := multiline.Read(ctx, editor)
		if err != nil {
			return err
		}
		fmt.Println("-----")
		for _, s := range lines {
			fmt.Println(s)
		}
		fmt.Println("-----")
	}
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
