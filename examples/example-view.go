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

func try(ctx context.Context, ed *multiline.Editor) error {
	lines, err := ed.Read(ctx)
	if err != nil {
		return err
	}
	fmt.Println("-----")
	fmt.Println(strings.Join(lines, "\n"))
	fmt.Println("-----")
	return nil
}

func mains() error {
	ctx := context.Background()
	var ed multiline.Editor

	ed.SetWriter(colorable.NewColorableStdout())
	lines := []string{}
	for i := 0; i < 50; i++ {
		lines = append(lines, fmt.Sprintf("LINE=%d", i))
	}
	ed.SetDefault(lines)

	fmt.Println("When .moveEnd=false")
	ed.SetMoveEnd(false)
	if err := try(ctx, &ed); err != nil {
		return err
	}
	fmt.Println("When .moveEnd=true")
	ed.SetMoveEnd(true)
	return try(ctx, &ed)
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
