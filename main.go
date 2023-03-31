package multiline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/nyaosorg/go-readline-ny"
)

type MultiLine struct {
	editor  readline.Editor
	csrline int
	lines   []string

	after         func(string) bool
	origBackSpace readline.KeyFuncT
	origDel       readline.KeyFuncT
}

func (m *MultiLine) updateLine(line string) {
	if m.csrline >= len(m.lines) {
		m.lines = append(m.lines, line)
	} else {
		m.lines[m.csrline] = line
	}
}

func (m *MultiLine) up(_ context.Context, _ *readline.Buffer) readline.Result {
	if m.csrline <= 0 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		m.updateLine(line)
		m.csrline--
		fmt.Fprint(m.editor.Out, "\r\x1B[A")
		return true
	}
	return readline.ENTER
}

func (m *MultiLine) submit(_ context.Context, B *readline.Buffer) readline.Result {
	fmt.Fprintln(m.editor.Out)
	for i := m.csrline + 1; i < len(m.lines); i++ {
		fmt.Fprintln(m.editor.Out)
	}
	m.editor.Out.Flush()
	m.after = func(line string) bool {
		m.updateLine(line)
		return false
	}
	return readline.ENTER
}

func (m *MultiLine) down(_ context.Context, _ *readline.Buffer) readline.Result {
	if m.csrline >= len(m.lines)-1 {
		return readline.CONTINUE
	}
	fmt.Fprintln(m.editor.Out)
	m.after = func(line string) bool {
		m.updateLine(line)
		m.csrline++
		return true
	}
	return readline.ENTER
}

func (m *MultiLine) joinAbove(ctx context.Context, b *readline.Buffer) readline.Result {
	if b.Cursor > 0 {
		return m.origBackSpace.Call(ctx, b)
	}
	if m.csrline == 0 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		if m.csrline > 0 {
			m.csrline--
			m.editor.Cursor = utf8.RuneCountInString(m.lines[m.csrline])
			m.lines[m.csrline] = m.lines[m.csrline] + line
			if m.csrline+1 < len(m.lines) {
				copy(m.lines[m.csrline+1:], m.lines[m.csrline+2:])
				m.lines = m.lines[:len(m.lines)-1]
			}
			io.WriteString(m.editor.Out, "\x1B[A\r\x1B[s")
			if m.csrline < len(m.lines) {
				i := m.csrline
				for {
					fmt.Fprintf(m.editor.Out, "%2d %s\x1B[0K", i+1, m.lines[i])
					i++
					if i >= len(m.lines) {
						break
					}
					fmt.Fprintln(m.editor.Out)
				}
			}
			io.WriteString(m.editor.Out, "\x1B[J\x1B[u")
		}
		return true
	}
	return readline.ENTER
}

func (m *MultiLine) newLine(_ context.Context, b *readline.Buffer) readline.Result {
	var sb strings.Builder
	for _, mm := range b.Buffer[b.Cursor:] {
		mm.Moji.WriteTo(&sb)
	}
	if m.csrline >= len(m.lines) {
		m.lines = append(m.lines, "")
	}
	m.lines = append(m.lines, "")
	copy(m.lines[m.csrline+2:], m.lines[m.csrline+1:])
	m.lines[m.csrline+1] = sb.String()
	b.Buffer = b.Buffer[:b.Cursor]
	b.RepaintAll()

	m.after = func(line string) bool {
		io.WriteString(m.editor.Out, "\x1B[K\n\x1B[s")
		m.updateLine(line)
		m.editor.Cursor = 0
		m.csrline++
		if m.csrline < len(m.lines) {
			i := m.csrline
			for {
				fmt.Fprintf(m.editor.Out, "%2d %s\x1B[0K", i+1, m.lines[i])
				i++
				if i >= len(m.lines) {
					break
				}
				fmt.Fprintln(m.editor.Out)
			}
		}
		io.WriteString(m.editor.Out, "\x1B[J\x1B[u")
		return true
	}
	return readline.ENTER
}

func (m *MultiLine) joinBelow(ctx context.Context, b *readline.Buffer) readline.Result {
	if len(b.Buffer) <= 0 {
		if len(m.lines) <= 0 {
			return m.origDel.Call(ctx, b)
		}
		if len(m.lines) == 1 && m.csrline == 0 {
			return m.origDel.Call(ctx, b)
		}
	}
	if b.Cursor < len(b.Buffer) {
		return m.origDel.Call(ctx, b)
	}
	if m.csrline+1 < len(m.lines) {
		b.InsertString(b.Cursor, m.lines[m.csrline+1])
		b.Out.WriteString("\x1B[s")
		copy(m.lines[m.csrline+1:], m.lines[m.csrline+2:])
		m.lines = m.lines[:len(m.lines)-1]
		for i := m.csrline + 1; i < len(m.lines); i++ {
			fmt.Fprintf(m.editor.Out, "\n%2d %s\x1B[K", i+1, m.lines[i])
		}
		b.Out.WriteString("\x1B[J\x1B[u")
		b.RepaintAll()
		b.Out.Flush()
	}
	return readline.CONTINUE
}

func (m *MultiLine) printAfter(i int) {
	if i < len(m.lines) {
		for {
			fmt.Fprintf(m.editor.Out, "%2d %s\x1B[K", i+1, m.lines[i])
			i++
			if i >= len(m.lines) {
				break
			}
			fmt.Fprintln(m.editor.Out)
		}
	}
	io.WriteString(m.editor.Out, "\x1B[J")
	m.editor.Out.Flush()
}

func (m *MultiLine) repaint(_ context.Context, b *readline.Buffer) readline.Result {
	io.WriteString(m.editor.Out, "\x1B[1;1H\x1B[2J")
	m.printAfter(0)
	if m.csrline < len(m.lines)-1 {
		fmt.Fprintf(m.editor.Out, "\x1B[%dA", len(m.lines)-1-m.csrline)
	}
	b.RepaintAll()
	return readline.CONTINUE
}

func New() *MultiLine {
	m := &MultiLine{}

	m.origDel = m.editor.GetBindKey(readline.K_CTRL_D)
	m.origBackSpace = m.editor.GetBindKey(readline.K_CTRL_H)
	m.editor.LineFeed = func(rc readline.Result) {
		if rc != readline.ENTER {
			fmt.Fprintln(m.editor.Out)
		}
	}
	m.editor.Prompt = func() (int, error) {
		return fmt.Fprintf(m.editor.Out, "%2d ", m.csrline+1)
	}
	m.editor.BindKeyClosure(readline.K_CTRL_D, m.joinBelow)
	m.editor.BindKeyClosure(readline.K_CTRL_H, m.joinAbove)
	m.editor.BindKeyClosure(readline.K_CTRL_J, m.submit)
	m.editor.BindKeyClosure(readline.K_CTRL_L, m.repaint)
	m.editor.BindKeyClosure(readline.K_CTRL_M, m.newLine)
	m.editor.BindKeyClosure(readline.K_CTRL_N, m.down)
	m.editor.BindKeyClosure(readline.K_CTRL_P, m.up)
	m.editor.BindKeyClosure(readline.K_DELETE, m.joinBelow)
	m.editor.BindKeyClosure(readline.K_DOWN, m.down)
	m.editor.BindKeyClosure(readline.K_UP, m.up)
	return m
}

func (m *MultiLine) Read(ctx context.Context) ([]string, error) {
	m.csrline = 0
	m.lines = []string{}

	for {
		if m.csrline < len(m.lines) {
			m.editor.Default = m.lines[m.csrline]
		} else {
			m.editor.Default = ""
		}
		m.after = func(string) bool { return true }
		line, err := m.editor.ReadLine(ctx)
		if err != nil {
			if errors.Is(err, readline.CtrlC) {
				m.lines = m.lines[:0]
				m.csrline = 0
				fmt.Fprintln(m.editor.Out, "^C")
				continue
			}
			return nil, err
		}
		m.editor.Out.Flush()
		if !m.after(line) {
			return m.lines, nil
		}
		m.editor.Out.Flush()
	}
}

func Read(ctx context.Context) ([]string, error) {
	return New().Read(ctx)
}
