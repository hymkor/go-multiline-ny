package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/nyaosorg/go-readline-ny"
)

func readMultiLine(ctx context.Context, editor *readline.Editor) ([]string, error) {
	submit := false
	editor.BindKeyClosure(readline.K_CTRL_J, func(_ context.Context, B *readline.Buffer) readline.Result {
		submit = true
		return readline.ENTER
	})

	csrline := 0
	upper := false
	upperFunc := func(_ context.Context, B *readline.Buffer) readline.Result {
		if csrline <= 0 {
			return readline.CONTINUE
		}
		upper = true
		return readline.ENTER
	}
	editor.BindKeyClosure(readline.K_CTRL_P, upperFunc)
	editor.BindKeyClosure(readline.K_UP, upperFunc)

	editor.LineFeed = func(rc readline.Result) {
		if rc == readline.ENTER && upper {
			return
		}
		fmt.Fprintln(editor.Out)
	}

	enterFunc, err := readline.GetFunc(readline.F_ACCEPT_LINE)
	if err != nil {
		return nil, err
	}
	editor.BindKeyFunc(readline.K_DOWN, enterFunc)
	editor.BindKeyFunc(readline.K_CTRL_N, enterFunc)

	lines := []string{}

	editor.Prompt = func() (int, error) {
		return fmt.Printf("%d> ", csrline)
	}
	for {
		if csrline < len(lines) {
			editor.Default = lines[csrline]
		} else {
			editor.Default = ""
		}
		line, err := editor.ReadLine(ctx)
		if err != nil {
			if errors.Is(err, readline.CtrlC) {
				lines = lines[:0]
				fmt.Println("^C")
				continue
			}
			return nil, err
		}
		if csrline >= len(lines) {
			lines = append(lines, line)
		} else {
			lines[csrline] = line
		}
		if submit {
			for i := csrline + 1; i < len(lines); i++ {
				fmt.Println()
			}
			return lines, nil
		} else if upper {
			upper = false
			csrline--
			fmt.Printf("\r\x1B[A")
		} else {
			csrline++
		}
	}
}

func mains() error {
	editor := &readline.Editor{}
	ctx := context.Background()
	fmt.Println("Enter, DOWN or Ctrl-N: New line or move to the next line")
	fmt.Println("UP or Ctrl-P: Move to the previous line.")
	fmt.Println("Ctrl-Enter: Sumbit")
	fmt.Println("Ctrl-C: Cancel lines.")
	fmt.Println("Ctrl-D: Quit.")
	for {
		lines, err := readMultiLine(ctx, editor)
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
