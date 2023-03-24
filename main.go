package multiline

import (
	"context"
	"errors"
	"fmt"

	"github.com/nyaosorg/go-readline-ny"
)

func Read(ctx context.Context) ([]string, error) {
	const (
		NEWLINE = iota
		COMMIT
		UP
	)

	press := NEWLINE

	editor := &readline.Editor{}

	editor.BindKeyClosure(readline.K_CTRL_J, func(_ context.Context, B *readline.Buffer) readline.Result {
		press = COMMIT
		return readline.ENTER
	})

	csrline := 0
	upperFunc := func(_ context.Context, B *readline.Buffer) readline.Result {
		if csrline <= 0 {
			return readline.CONTINUE
		}
		press = UP
		return readline.ENTER
	}
	editor.BindKeyClosure(readline.K_CTRL_P, upperFunc)
	editor.BindKeyClosure(readline.K_UP, upperFunc)

	editor.LineFeed = func(rc readline.Result) {
		if rc == readline.ENTER && press == UP {
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
		return fmt.Fprintf(editor.Out, "%2d ", csrline+1)
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
				csrline = 0
				fmt.Fprintln(editor.Out, "^C")
				continue
			}
			return nil, err
		}
		if csrline >= len(lines) {
			lines = append(lines, line)
		} else {
			lines[csrline] = line
		}
		if press == COMMIT {
			for i := csrline + 1; i < len(lines); i++ {
				fmt.Fprintln(editor.Out)
			}
			editor.Out.Flush()
			return lines, nil
		} else if press == UP {
			press = NEWLINE
			csrline--
			fmt.Fprint(editor.Out, "\r\x1B[A")
		} else {
			csrline++
		}
	}
}
