package multiline

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-runewidth"

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
	historyPtr    int

	viewWidth int

	Prompt func(w io.Writer, i int) (int, error)
}

func (m *Editor) SetColoring(c readline.Coloring)               { m.LineEditor.Coloring = c }
func (m *Editor) SetHistory(h readline.IHistory)                { m.LineEditor.History = h }
func (m *Editor) SetPrompt(f func(io.Writer, int) (int, error)) { m.Prompt = f }
func (m *Editor) SetWriter(w io.Writer)                         { m.LineEditor.Writer = w }

func (m *Editor) storeCurrentLine(line string) {
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
		m.storeCurrentLine(line)
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
		m.storeCurrentLine(line)
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
		m.storeCurrentLine(line)
		m.csrline++
		return true
	}
	return readline.ENTER
}

func mojiCount(s string) int {
	return len(readline.StringToMoji(s))
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
			m.LineEditor.Cursor = mojiCount(m.lines[m.csrline])
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
	// make new line at the next of the cursor
	if m.csrline >= len(m.lines) {
		m.lines = append(m.lines, "")
	}
	m.lines = append(m.lines, "")
	copy(m.lines[m.csrline+2:], m.lines[m.csrline+1:])

	// move characters after cursor to the nextline
	m.lines[m.csrline+1] = b.SubString(b.Cursor, len(b.Buffer))
	b.Buffer = b.Buffer[:b.Cursor]

	b.RepaintAll()

	m.after = func(line string) bool {
		io.WriteString(m.LineEditor.Out, "\x1B[K\n")
		m.storeCurrentLine(line)
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
			m.printOne(i)
		}
		b.Out.WriteString("\x1B[J\x1B[u")
		b.RepaintAll()
		b.Out.Flush()
	}
	return readline.CONTINUE
}

const forbiddenWidth = 3

func (m *Editor) printOne(i int) {
	w, _ := m.Prompt(m.LineEditor.Out, i)
	defaultColor := m.LineEditor.Coloring.Init()
	color := defaultColor
	for _, c := range m.lines[i] {
		newColor := m.LineEditor.Coloring.Next(c)
		if newColor != color {
			newColor.WriteTo(m.LineEditor.Out)
		}
		color = newColor
		if c < 0x20 {
			if w+2 >= m.viewWidth-forbiddenWidth {
				break
			}
			m.LineEditor.Out.Write([]byte{'^', '@' + byte(c)})
			w += 2
		} else {
			w1 := runewidth.RuneWidth(c)
			if w+w1 >= m.viewWidth-forbiddenWidth {
				break
			}
			m.LineEditor.Out.WriteRune(c)
			w += w1
		}
	}
	if color != defaultColor {
		defaultColor.WriteTo(m.LineEditor.Out)
	}
	io.WriteString(m.LineEditor.Out, "\x1B[K")
}

func (m *Editor) printFromTo(i, j int) int {
	lfCount := 0
	if i < j {
		for {
			m.printOne(i)
			i++
			if i >= j {
				break
			}
			fmt.Fprintln(m.LineEditor.Out)
			lfCount++
		}
	}
	return lfCount
}

func (m *Editor) printAfter(i int) int {
	lfCount := m.printFromTo(i, len(m.lines))
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

func (m *Editor) printCurrentHistoryRecord(string) bool {
	// clear
	if m.csrline > 0 {
		fmt.Fprintf(m.LineEditor.Out, "\x1B[%dA", m.csrline)
	}
	io.WriteString(m.LineEditor.Out, "\r")

	m.lines = strings.Split(m.LineEditor.History.At(m.historyPtr), "\n")
	m.csrline = 0
	for m.csrline < len(m.lines)-1 {
		m.printOne(m.csrline)
		fmt.Fprintln(m.LineEditor.Out)
		m.csrline++
	}
	fmt.Fprint(m.LineEditor.Out, "\x1B[J")
	m.LineEditor.Cursor = 9999
	return true
}

func (m *Editor) prevHistory(_ context.Context, b *readline.Buffer) readline.Result {
	if m.LineEditor.History == nil || m.LineEditor.History.Len() <= 0 {
		return readline.CONTINUE
	}
	if m.historyPtr <= 0 {
		if !m.LineEditor.HistoryCycling {
			return readline.CONTINUE
		}
		m.historyPtr = m.LineEditor.History.Len()
	}
	m.historyPtr--
	m.after = m.printCurrentHistoryRecord
	return readline.ENTER
}

func (m *Editor) nextHistory(_ context.Context, b *readline.Buffer) readline.Result {
	if m.LineEditor.History == nil || m.LineEditor.History.Len() <= 0 {
		return readline.CONTINUE
	}
	if m.historyPtr+1 >= m.LineEditor.History.Len() {
		if !m.LineEditor.HistoryCycling {
			return readline.CONTINUE
		}
		m.historyPtr = -1
	}
	m.historyPtr++
	m.after = m.printCurrentHistoryRecord
	return readline.ENTER
}

func (m *Editor) clear(_ context.Context, b *readline.Buffer) readline.Result {
	m.after = func(string) bool {
		if m.csrline > 0 {
			fmt.Fprintf(m.LineEditor.Out, "\x1B[%dA", m.csrline)
		}
		io.WriteString(m.LineEditor.Out, "\r\x1B[J")
		m.csrline = 0
		m.lines = m.lines[:0]
		return true
	}
	return readline.ENTER
}

func insertSliceAt(slice []string, pos int, newlines []string) []string {
	backup := make([]string, len(slice)-pos)
	copy(backup, slice[pos:])
	slice = slice[:pos]
	slice = append(slice, newlines...)
	slice = append(slice, backup...)
	return slice
}

func (m *Editor) paste(_ context.Context, b *readline.Buffer) readline.Result {
	text, err := clipboard.ReadAll()
	if err != nil {
		return readline.CONTINUE
	}
	text = strings.TrimRight(text, "\r\n\000")
	if len(text) <= 0 {
		return readline.CONTINUE
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	newlines := strings.Split(text, "\n")
	if len(newlines) <= 0 {
		return readline.CONTINUE
	}
	if len(newlines) <= 1 {
		b.InsertAndRepaint(newlines[0])
		return readline.CONTINUE
	}

	tmp := b.SubString(b.Cursor, len(b.Buffer))
	nextCursorPosition := mojiCount(tmp)
	b.Buffer = b.Buffer[:b.Cursor]
	b.InsertAndRepaint(newlines[0])
	b.Out.Flush()

	newlines = newlines[1:]
	newlines[len(newlines)-1] += tmp

	m.after = func(line string) bool {
		m.storeCurrentLine(line)
		fmt.Fprintln(m.LineEditor.Out)
		m.csrline++

		m.lines = insertSliceAt(m.lines, m.csrline, newlines)

		m.printFromTo(m.csrline, m.csrline+len(newlines))
		if m.csrline+len(newlines) < len(m.lines) {
			fmt.Fprintln(m.LineEditor.Out)
			lfCount := 1 + m.printAfter(m.csrline+len(newlines))
			fmt.Fprintf(m.LineEditor.Out, "\x1B[%dA", lfCount)
		}
		m.LineEditor.Out.Flush()
		m.csrline += len(newlines) - 1
		m.LineEditor.Cursor = mojiCount(m.lines[m.csrline]) - nextCursorPosition
		return true
	}
	return readline.ENTER
}

func (m *Editor) init() error {
	tty1, err := readline.NewDefaultTty()
	if err != nil {
		return err
	}
	m.viewWidth, _, err = tty1.Size()
	tty1.Close()
	if err != nil {
		return err
	}

	m.inited = true
	m.origDel = m.LineEditor.GetBindKey(readline.K_CTRL_D)
	m.origBackSpace = m.LineEditor.GetBindKey(readline.K_CTRL_H)
	m.LineEditor.LineFeed = func(rc readline.Result) {
		if rc != readline.ENTER {
			fmt.Fprintln(m.LineEditor.Out)
		}
	}
	if m.Prompt == nil {
		m.Prompt = func(w io.Writer, i int) (int, error) {
			return fmt.Fprintf(w, "%2d ", i+1)
		}
	}
	m.LineEditor.Prompt = func() (int, error) {
		return m.Prompt(m.LineEditor.Out, m.csrline)
	}

	m.LineEditor.BindKeyClosure(readline.K_ALT_N, m.nextHistory)
	m.LineEditor.BindKeyClosure(readline.K_ALT_P, m.prevHistory)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_D, m.joinBelow)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_DOWN, m.nextHistory)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_H, m.joinAbove)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_J, m.submit)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_L, m.repaint)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_M, m.newLine)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_N, m.down)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_P, m.up)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_UP, m.prevHistory)
	m.LineEditor.BindKeyClosure(readline.K_CTRL_Y, m.paste)
	m.LineEditor.BindKeyClosure(readline.K_DELETE, m.joinBelow)
	m.LineEditor.BindKeyClosure(readline.K_DOWN, m.down)
	m.LineEditor.BindKeyClosure(readline.K_ESCAPE, m.clear)
	m.LineEditor.BindKeyClosure(readline.K_PAGEDOWN, m.nextHistory)
	m.LineEditor.BindKeyClosure(readline.K_PAGEUP, m.prevHistory)
	m.LineEditor.BindKeyClosure(readline.K_UP, m.up)

	return nil
}

// New is same as `&multiline.Editor{}`.
// It exists only for compatibility.
// You can use `var m multiline.Editor`.
func New() *Editor {
	return &Editor{}
}

func (m *Editor) Read(ctx context.Context) ([]string, error) {
	if !m.inited {
		if err := m.init(); err != nil {
			return nil, err
		}
	}
	m.csrline = 0
	m.lines = []string{}
	if m.LineEditor.History != nil {
		m.historyPtr = m.LineEditor.History.Len()
	}

	for {
		if m.csrline < len(m.lines) {
			m.LineEditor.Default = m.lines[m.csrline]
		} else {
			m.LineEditor.Default = ""
		}
		m.after = func(string) bool { return true }
		line, err := m.LineEditor.ReadLine(ctx)
		if err != nil {
			return nil, err
		}
		m.LineEditor.Out.Flush()
		if !m.after(line) {
			return m.lines, nil
		}
		m.LineEditor.Out.Flush()
	}
}

// Read (The function version) is same as `var m multiline.Editor ; m.Read(context)`.
// It Exists only for compatibility.
func Read(ctx context.Context) ([]string, error) {
	var m Editor
	return m.Read(ctx)
}
