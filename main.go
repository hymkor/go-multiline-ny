package multiline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nyaosorg/go-readline-ny"
)

func Read(ctx context.Context) ([]string, error) {
	const (
		NEWLINE = iota
		COMMIT
		UP
		DOWN
		JOINBEFORE
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
			if press == UP || press == JOINBEFORE {
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

	bs := editor.GetBindKey(readline.K_CTRL_H)
	joinbefore := func(ctx context.Context, b *readline.Buffer) readline.Result {
		if b.Cursor > 0 {
			return bs.Call(ctx, b)
		}
		if csrline == 0 {
			return readline.CONTINUE
		}
		press = JOINBEFORE
		return readline.ENTER
	}
	editor.BindKeyClosure(readline.K_CTRL_H, joinbefore)
	lines := []string{}

	editor.BindKeyClosure(readline.K_CTRL_M, func(_ context.Context, b *readline.Buffer) readline.Result {
		var sb strings.Builder
		for _, m := range b.Buffer[b.Cursor:] {
			m.Moji.WriteTo(&sb)
		}
		if csrline >= len(lines) {
			lines = append(lines, "")
		}
		lines = append(lines, "")
		copy(lines[csrline+2:], lines[csrline+1:])
		lines[csrline+1] = sb.String()
		b.Buffer = b.Buffer[:b.Cursor]
		b.RepaintAll()
		return readline.ENTER
	})

	del := editor.GetBindKey(readline.K_CTRL_D)
	joinafter := func(ctx context.Context, b *readline.Buffer) readline.Result {
		if len(b.Buffer) <= 0 {
			if len(lines) <= 0 {
				return del.Call(ctx, b)
			}
			if len(lines) == 1 && csrline == 0 {
				return del.Call(ctx, b)
			}
		}
		if b.Cursor < len(b.Buffer) {
			return del.Call(ctx, b)
		}
		if csrline+1 < len(lines) {
			b.InsertString(b.Cursor, lines[csrline+1])
			b.Out.WriteString("\x1B[s")
			copy(lines[csrline+1:], lines[csrline+2:])
			lines = lines[:len(lines)-1]
			for i := csrline + 1; i < len(lines); i++ {
				fmt.Fprintf(editor.Out, "\n%2d %s\x1B[K", i, lines[i])
			}
			b.Out.WriteString("\x1B[J\x1B[u")
			b.RepaintAll()
			b.Out.Flush()
		}
		return readline.CONTINUE
	}
	editor.BindKeyClosure(readline.K_CTRL_D, joinafter)
	editor.BindKeyClosure(readline.K_DELETE, joinafter)

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
			if csrline >= len(lines) {
				lines = append(lines, line)
			} else {
				lines[csrline] = line
			}
			csrline++
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
		} else if press == JOINBEFORE {
			if csrline > 0 {
				csrline--
				editor.Cursor = utf8.RuneCountInString(lines[csrline])
				lines[csrline] = lines[csrline] + line
				if csrline+1 < len(lines) {
					copy(lines[csrline+1:], lines[csrline+2:])
					lines = lines[:len(lines)-1]
				}
				fmt.Fprint(editor.Out, "\x1B[A\r")
				for i := csrline; i < len(lines); i++ {
					fmt.Fprintf(editor.Out, "%2d %s\x1B[0K\n", i+1, lines[i])
				}
				if len(lines) > csrline {
					fmt.Fprintf(editor.Out, "\x1B[0K\x1B[%dA", len(lines)-csrline)
				}
			}
			press = NEWLINE
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
