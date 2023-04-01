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

type Editor struct {
	LineEditor readline.Editor
	csrline    int
	lines      []string
	inited     bool

	after         func(string) bool
	origBackSpace readline.KeyFuncT
	origDel       readline.KeyFuncT

	Prompt func(w io.Writer, i int) (int, error)
}

func (m *Editor) updateLine(line string) {
	if m.csrline >= len(m.lines) {
		m.lines = append(m.lines, line)
	} else {
		m.lines[m.csrline] = line
	}
}

func (m *Editor) up(_ context.Context, _ *readline.Buffer) readline.Result {
	if m.csrline <= 0 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		m.updateLine(line)
		m.csrline--
		fmt.Fprint(m.LineEditor.Out, "\r\x1B[A")
		return true
	}
	return readline.ENTER
}

func (m *Editor) submit(_ context.Context, B *readline.Buffer) readline.Result {
	fmt.Fprintln(m.LineEditor.Out)
	for i := m.csrline + 1; i < len(m.lines); i++ {
		fmt.Fprintln(m.LineEditor.Out)
	}
	m.LineEditor.Out.Flush()
	m.after = func(line string) bool {
		m.updateLine(line)
		return false
	}
	return readline.ENTER
}

func (m *Editor) down(_ context.Context, _ *readline.Buffer) readline.Result {
	if m.csrline >= len(m.lines)-1 {
		return readline.CONTINUE
	}
	fmt.Fprintln(m.LineEditor.Out)
	m.after = func(line string) bool {
		m.updateLine(line)
		m.csrline++
		return true
	}
	return readline.ENTER
}

func (m *Editor) joinAbove(ctx context.Context, b *readline.Buffer) readline.Result {
	if b.Cursor > 0 {
		return m.origBackSpace.Call(ctx, b)
	}
	if m.csrline == 0 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		if m.csrline > 0 {
			m.csrline--
			m.LineEditor.Cursor = utf8.RuneCountInString(m.lines[m.csrline])
			m.lines[m.csrline] = m.lines[m.csrline] + line
			if m.csrline+1 < len(m.lines) {
				copy(m.lines[m.csrline+1:], m.lines[m.csrline+2:])
				m.lines = m.lines[:len(m.lines)-1]
			}
			io.WriteString(m.LineEditor.Out, "\x1B[A\r\x1B[s")
			m.printAfter(m.csrline)
			io.WriteString(m.LineEditor.Out, "\x1B[u")
		}
		return true
	}
	return readline.ENTER
}

func (m *Editor) newLine(_ context.Context, b *readline.Buffer) readline.Result {
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
		io.WriteString(m.LineEditor.Out, "\x1B[K\n")
		m.updateLine(line)
		m.LineEditor.Cursor = 0
		m.csrline++
		lfCount := m.printAfter(m.csrline)
		if lfCount > 0 {
			fmt.Fprintf(m.LineEditor.Out, "\x1B[%dA", lfCount)
		}
		return true
	}
	return readline.ENTER
}

func (m *Editor) joinBelow(ctx context.Context, b *readline.Buffer) readline.Result {
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
			fmt.Fprintln(m.LineEditor.Out)
			m.Prompt(m.LineEditor.Out, i)
			fmt.Fprintf(m.LineEditor.Out, "%s\x1B[K", m.lines[i])
		}
		b.Out.WriteString("\x1B[J\x1B[u")
		b.RepaintAll()
		b.Out.Flush()
	}
	return readline.CONTINUE
}

func (m *Editor) printAfter(i int) int {
	lfCount := 0
	if i < len(m.lines) {
		for {
			m.Prompt(m.LineEditor.Out, i)
			fmt.Fprintf(m.LineEditor.Out, "%s\x1B[K", m.lines[i])
			i++
			if i >= len(m.lines) {
				break
			}
			fmt.Fprintln(m.LineEditor.Out)
			lfCount++
		}
	}
	io.WriteString(m.LineEditor.Out, "\x1B[J")
	m.LineEditor.Out.Flush()
	return lfCount
}

func (m *Editor) repaint(_ context.Context, b *readline.Buffer) readline.Result {
	io.WriteString(m.LineEditor.Out, "\x1B[1;1H\x1B[2J")
	m.printAfter(0)
	if m.csrline < len(m.lines)-1 {
		fmt.Fprintf(m.LineEditor.Out, "\x1B[%dA", len(m.lines)-1-m.csrline)
	}
	b.RepaintAll()
	return readline.CONTINUE
}

func (m *Editor) init() {
	m.inited = true
	m.origDel = m.LineEditor.GetBindKey(readline.K_CTRL_D)
	m.origBackSpace = m.LineEditor.GetBindKey(readline.K_CTRL_H)
	m.LineEditor.LineFeed = func(rc readline.Result) {
		if rc != readline.ENTER {
			fmt.Fprintln(m.LineEditor.Out)
		}
	}
	m.Prompt = func(w io.Writer, i int) (int, error) {
		return fmt.Fprintf(w, "%2d ", i+1)
	}
	m.LineEditor.Prompt = func() (int, error) {
		return m.Prompt(m.LineEditor.Out, m.csrline)
	}
	m.LineEditor.BindKeyClosure(readline.K_CTRL_D, m.joinBelow)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_H, m.joinAbove)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_J, m.submit)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_L, m.repaint)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_M, m.newLine)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_N, m.down)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_P, m.up)
	m.LineEditor.BindKeyClosure(readline.K_DELETE, m.joinBelow)
	m.LineEditor.BindKeyClosure(readline.K_DOWN, m.down)
	m.LineEditor.BindKeyClosure(readline.K_UP, m.up)
}

func New() *Editor {
	m := &Editor{}
	m.init()
	return m
}

func (m *Editor) Read(ctx context.Context) ([]string, error) {
	if !m.inited {
		m.init()
	}
	m.csrline = 0
	m.lines = []string{}

	for {
		if m.csrline < len(m.lines) {
			m.LineEditor.Default = m.lines[m.csrline]
		} else {
			m.LineEditor.Default = ""
		}
		m.after = func(string) bool { return true }
		line, err := m.LineEditor.ReadLine(ctx)
		if err != nil {
			if errors.Is(err, readline.CtrlC) {
				m.lines = m.lines[:0]
				m.csrline = 0
				fmt.Fprintln(m.LineEditor.Out, "^C")
				continue
			}
			return nil, err
		}
		m.LineEditor.Out.Flush()
		if !m.after(line) {
			return m.lines, nil
		}
		m.LineEditor.Out.Flush()
	}
}

func Read(ctx context.Context) ([]string, error) {
	var m Editor
	return m.Read(ctx)
}
