package multiline

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-runewidth"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/keys"
)

type Editor struct {
	LineEditor readline.Editor
	csrline    int
	lines      []string
	after      func(string) bool
	historyPtr int
	viewWidth  int // when viewWidth==0, it means the instance is not initialized, yet
	prompt     func(w io.Writer, i int) (int, error)
}

func (m *Editor) SetHistoryCycling(value bool)                  { m.LineEditor.HistoryCycling = value }
func (m *Editor) SetColoring(c readline.Coloring)               { m.LineEditor.Coloring = c }
func (m *Editor) SetHistory(h readline.IHistory)                { m.LineEditor.History = h }
func (m *Editor) SetPrompt(f func(io.Writer, int) (int, error)) { m.prompt = f }
func (m *Editor) SetWriter(w io.Writer)                         { m.LineEditor.Writer = w }

func (m *Editor) SwapEnter() error {
	m.BindKey(keys.CtrlM, readline.AnonymousCommand(m.submit))
	m.BindKey(keys.CtrlJ, readline.AnonymousCommand(m.newLine))
	return nil
}

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

func (m *Editor) left(ctx context.Context, b *readline.Buffer) readline.Result {
	if b.Cursor > 0 {
		return readline.CmdBackwardChar.Call(ctx, b)
	}
	if m.csrline == 0 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		m.storeCurrentLine(line)
		m.csrline--
		m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline])
		fmt.Fprint(m.LineEditor.Out, "\r\x1B[A")
		return true
	}
	return readline.ENTER
}

func (m *Editor) right(ctx context.Context, b *readline.Buffer) readline.Result {
	if b.Cursor < len(b.Buffer) {
		return readline.CmdForwardChar.Call(ctx, b)
	}
	if m.csrline+1 >= len(m.lines) {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		m.storeCurrentLine(line)
		m.csrline++
		m.LineEditor.Cursor = 0
		fmt.Fprint(m.LineEditor.Out, "\n")
		return true
	}
	return readline.ENTER
}

func (m *Editor) joinAbove(ctx context.Context, b *readline.Buffer) readline.Result {
	if b.Cursor > 0 {
		return readline.CmdBackwardDeleteChar.Call(ctx, b)
	}
	if m.csrline == 0 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		if m.csrline > 0 {
			m.csrline--
			m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline])
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
			return readline.CmdDeleteOrAbort.Call(ctx, b)
		}
		if len(m.lines) == 1 && m.csrline == 0 {
			return readline.CmdDeleteOrAbort.Call(ctx, b)
		}
	}
	if b.Cursor < len(b.Buffer) {
		return readline.CmdDeleteOrAbort.Call(ctx, b)
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
	w, _ := m.prompt(m.LineEditor.Out, i)
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
	nextCursorPosition := len(b.Buffer) - b.Cursor
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
		m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline]) - nextCursorPosition
		return true
	}
	return readline.ENTER
}

func (m *Editor) init() error {
	if m.viewWidth > 0 {
		return nil
	}
	var err error
	m.viewWidth, _, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	m.LineEditor.LineFeed = func(rc readline.Result) {
		if rc != readline.ENTER {
			fmt.Fprintln(m.LineEditor.Out)
		}
	}
	if m.prompt == nil {
		m.prompt = func(w io.Writer, i int) (int, error) {
			return fmt.Fprintf(w, "%2d ", i+1)
		}
	}
	m.LineEditor.PromptWriter = func(w io.Writer) (int, error) {
		return m.prompt(w, m.csrline)
	}

	type ac = readline.AnonymousCommand

	m.LineEditor.BindKey(keys.AltN, ac(m.nextHistory))
	m.LineEditor.BindKey(keys.AltP, ac(m.prevHistory))
	m.LineEditor.BindKey(keys.CtrlB, ac(m.left))
	m.LineEditor.BindKey(keys.CtrlD, ac(m.joinBelow))
	m.LineEditor.BindKey(keys.CtrlDown, ac(m.nextHistory))
	m.LineEditor.BindKey(keys.CtrlF, ac(m.right))
	m.LineEditor.BindKey(keys.CtrlH, ac(m.joinAbove))
	m.LineEditor.BindKey(keys.CtrlL, ac(m.repaint))
	m.LineEditor.BindKey(keys.CtrlN, ac(m.down))
	m.LineEditor.BindKey(keys.CtrlP, ac(m.up))
	m.LineEditor.BindKey(keys.CtrlUp, ac(m.prevHistory))
	m.LineEditor.BindKey(keys.CtrlY, ac(m.paste))
	m.LineEditor.BindKey(keys.Delete, ac(m.joinBelow))
	m.LineEditor.BindKey(keys.Down, ac(m.down))
	m.LineEditor.BindKey(keys.Escape, ac(m.clear))
	m.LineEditor.BindKey(keys.Left, ac(m.left))
	m.LineEditor.BindKey(keys.PageDown, ac(m.nextHistory))
	m.LineEditor.BindKey(keys.PageUp, ac(m.prevHistory))
	m.LineEditor.BindKey(keys.Right, ac(m.right))
	m.LineEditor.BindKey(keys.Up, ac(m.up))
	m.LineEditor.BindKey(keys.CtrlM, ac(m.newLine))
	m.LineEditor.BindKey(keys.CtrlJ, ac(m.submit))
	m.LineEditor.BindKey(keys.CtrlR, readline.SelfInserter(keys.CtrlR))
	m.LineEditor.BindKey(keys.CtrlS, readline.SelfInserter(keys.CtrlS))
	return nil
}

func (m *Editor) BindKey(key keys.Code, f readline.Command) error {
	if err := m.init(); err != nil {
		return err
	}
	m.LineEditor.BindKey(key, f)
	return nil
}

func (m *Editor) Read(ctx context.Context) ([]string, error) {
	if err := m.init(); err != nil {
		return nil, err
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
