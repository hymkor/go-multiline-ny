package multiline

import (
	"context"
	"errors"
	"fmt"

	"github.com/nyaosorg/go-readline-ny"
)

func Read(ctx context.Context, editor *readline.Editor) ([]string, error) {
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
