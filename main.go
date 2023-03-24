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
		DOWN
	)

	press := NEWLINE

	editor := &readline.Editor{}

	editor.BindKeyClosure(readline.K_CTRL_J, func(_ context.Context, B *readline.Buffer) readline.Result {
		press = COMMIT
		return readline.ENTER
	})

	csrline := 0
	upFunc := func(_ context.Context, _ *readline.Buffer) readline.Result {
		if csrline <= 0 {
			return readline.CONTINUE
		}
		press = UP
		return readline.ENTER
	}
	editor.BindKeyClosure(readline.K_CTRL_P, upFunc)
	editor.BindKeyClosure(readline.K_UP, upFunc)

	editor.LineFeed = func(rc readline.Result) {
		if rc == readline.ENTER {
			if press == UP {
				return
			} else if press == NEWLINE {
				fmt.Fprintln(editor.Out, "\x1B[0K")
				return
			}
		}
		fmt.Fprintln(editor.Out)
	}
	downFunc := func(_ context.Context, _ *readline.Buffer) readline.Result {
		press = DOWN
		return readline.ENTER
	}
	editor.BindKeyClosure(readline.K_DOWN, downFunc)
	editor.BindKeyClosure(readline.K_CTRL_N, downFunc)

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
		if press == NEWLINE {
			tmp := []rune(line)
			nextline := string(tmp[editor.Cursor:])
			line = string(tmp[:editor.Cursor])
			if csrline >= len(lines) {
				lines = append(lines, line)
			} else {
				lines[csrline] = line
			}
			csrline++
			lines = append(lines, "")
			copy(lines[csrline+1:], lines[csrline:])
			lines[csrline] = nextline
			editor.Cursor = 0

			up := 0
			for i := csrline; ; {
				fmt.Fprintf(editor.Out, "%2d %s\x1B[0K", i+1, lines[i])
				i++
				if i >= len(lines) {
					fmt.Fprint(editor.Out, "\r")
					break
				}
				fmt.Fprintln(editor.Out)
				up++
			}
			if up > 0 {
				fmt.Fprintf(editor.Out, "\x1B[%dA", up)
			}
		} else {
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
			} else if press == DOWN {
				press = NEWLINE
				csrline++
			}
		}
		editor.Out.Flush()
	}
}
