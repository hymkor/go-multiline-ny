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
	viewHeight int
	headline   int // the first line on the screen
	prompt     func(w io.Writer, i int) (int, error)
	defaults   []string
	moveEnd    bool
}

func (m *Editor) SetHistoryCycling(value bool)                  { m.LineEditor.HistoryCycling = value }
func (m *Editor) SetColoring(c readline.Coloring)               { m.LineEditor.Coloring = c }
func (m *Editor) SetHistory(h readline.IHistory)                { m.LineEditor.History = h }
func (m *Editor) SetPrompt(f func(io.Writer, int) (int, error)) { m.prompt = f }
func (m *Editor) SetWriter(w io.Writer)                         { m.LineEditor.Writer = w }
func (m *Editor) SetDefault(d []string)                         { m.defaults = d }
func (m *Editor) SetMoveEnd(value bool)                         { m.moveEnd = value }
func (m *Editor) CursorLine() int                               { return m.csrline }
func (m *Editor) Lines() []string                               { return m.lines }

// Deprecated:
func (m *Editor) SwapEnter() error {
	m.BindKey(keys.CtrlM, readline.AnonymousCommand(m.Submit))
	m.BindKey(keys.CtrlJ, readline.AnonymousCommand(m.NewLine))
	return nil
}

func (m *Editor) storeCurrentLine(line string) {
	if m.csrline >= len(m.lines) {
		m.lines = append(m.lines, line)
	} else {
		m.lines[m.csrline] = line
	}
}

func (m *Editor) escA(n int) {
	if n == 1 {
		io.WriteString(m.LineEditor.Out, "\r\x1B[A")
	} else if n > 0 {
		fmt.Fprintf(m.LineEditor.Out, "\r\x1B[%dA", n)
	}
}

func (m *Editor) up(_ context.Context, _ *readline.Buffer) readline.Result {
	if m.csrline <= 0 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		m.storeCurrentLine(line)
		m.csrline--
		m.LineEditor.Out.WriteByte('\r')
		if m.fixView() < 0 {
			m.escA(m.printAfter(m.csrline))
		} else {
			fmt.Fprint(m.LineEditor.Out, "\x1B[A")
		}
		return true
	}
	return readline.ENTER
}

func (m *Editor) GotoEndLine() func() {
	end := min(len(m.lines), m.headline+m.viewHeight)
	lfCount := 0
	for i := m.csrline; i < end; i++ {
		fmt.Fprintln(m.LineEditor.Out)
		lfCount++
	}
	m.LineEditor.Out.Flush()
	return func() {
		if lfCount > 0 {
			fmt.Fprintf(m.LineEditor.Out, "\x1B[2K\x1B[%dF", lfCount)
		}
	}
}

func (m *Editor) Submit(_ context.Context, B *readline.Buffer) readline.Result {
	m.after = func(line string) bool {
		m.storeCurrentLine(line)
		m.GotoEndLine()
		return false
	}
	return readline.ENTER
}

func (m *Editor) CmdNextLine(_ context.Context, _ *readline.Buffer) readline.Result {
	if m.csrline >= len(m.lines)-1 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		m.storeCurrentLine(line)
		m.csrline++
		if m.fixView() > 0 {
			m.escA(m.csrline - m.headline)
			m.printAfter(m.headline)
		} else {
			fmt.Fprintln(m.LineEditor.Out)
		}
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
		if m.fixView() < 0 {
			m.escA(m.printAfter(m.csrline))
		} else {
			fmt.Fprint(m.LineEditor.Out, "\r\x1B[A")
		}
		m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline])
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
		if m.fixView() > 0 {
			m.escA(m.csrline - m.headline)
			m.printAfter(m.headline)
		} else {
			fmt.Fprint(m.LineEditor.Out, "\n")
		}
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
			m.fixView()
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

func (m *Editor) NewLine(_ context.Context, b *readline.Buffer) readline.Result {
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
		m.fixView()
		m.escA(m.printAfter(m.csrline))
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

func cutEscapeSequenceAndOldLine(s string) string {
	var buffer strings.Builder
	esc := false
	for i, end := 0, len(s); i < end; i++ {
		r := s[i]
		switch r {
		case '\r', '\n':
			buffer.Reset()
		case '\x1B':
			esc = true
		default:
			if esc {
				if ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z') {
					esc = false
				}
			} else {
				buffer.WriteByte(r)
			}
		}
	}
	return buffer.String()
}

func (m *Editor) printOne(i int) {
	var buffer strings.Builder
	m.prompt(&buffer, i)
	promptStr := buffer.String()

	io.WriteString(m.LineEditor.Out, promptStr)
	w0 := int(readline.GetStringWidth(cutEscapeSequenceAndOldLine(promptStr)))
	w := w0
	defaultColor := m.LineEditor.Coloring.Init()
	color := defaultColor
	for _, c := range m.lines[i] {
		newColor := m.LineEditor.Coloring.Next(c)
		if newColor != color {
			newColor.WriteTo(m.LineEditor.Out)
		}
		color = newColor
		if c == '\t' {
			size := 4 - (w-w0)%4
			if w+size >= m.viewWidth-forbiddenWidth {
				break
			}
			io.WriteString(m.LineEditor.Out, "    "[:size])
			w += size
		} else if c < 0x20 {
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

// printAfter prints lines[i:j].
// `headline` must be corrected.
// It does not fix view.
func (m *Editor) printFromTo(i, j int) int {
	lfCount := 0
	if j > m.headline+m.viewHeight {
		j = m.headline + m.viewHeight
	}
	if i < j {
		m.LineEditor.Out.WriteByte('\r')
		for {
			m.printOne(i)
			i++
			if i >= j {
				break
			}
			m.LineEditor.Out.WriteByte('\n')
			lfCount++
		}
	}
	return lfCount
}

// printAfter prints lines[i:].
// `headline` must be corrected.
// It does not fix view.
func (m *Editor) printAfter(i int) int {
	lfCount := m.printFromTo(i, len(m.lines))
	io.WriteString(m.LineEditor.Out, "\x1B[J\r")
	m.LineEditor.Out.Flush()
	return lfCount
}

func (m *Editor) repaint(_ context.Context, b *readline.Buffer) readline.Result {
	io.WriteString(m.LineEditor.Out, "\x1B[1;1H\x1B[2J")
	lfCount := m.printAfter(m.headline)
	lfCount -= (m.csrline - m.headline)
	m.escA(lfCount)
	b.RepaintAll()
	return readline.CONTINUE
}

func (m *Editor) fixView() int {
	if m.csrline >= m.headline+m.viewHeight {
		m.headline = m.csrline - m.viewHeight + 1
		return +1
	} else if m.csrline < m.headline {
		m.headline = m.csrline
		return -1
	}
	return 0
}

func (m *Editor) printCurrentHistoryRecord(string) bool {
	// clear
	m.escA(m.csrline - m.headline)
	m.lines = strings.Split(m.LineEditor.History.At(m.historyPtr), "\n")
	m.csrline = len(m.lines) - 1
	m.fixView()
	lfCount := m.printAfter(m.headline)
	lfCount -= (m.csrline - m.headline)
	m.escA(lfCount)
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

func insertSliceAt(slice []string, pos int, newlines []string) []string {
	backup := make([]string, len(slice)-pos)
	copy(backup, slice[pos:])
	slice = slice[:pos]
	slice = append(slice, newlines...)
	slice = append(slice, backup...)
	return slice
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
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
		m.fixView()

		m.lines = insertSliceAt(m.lines, m.csrline, newlines)
		start := m.csrline
		m.csrline += len(newlines) - 1
		m.fixView()
		m.printAfter(start)
		m.escA(min(len(m.lines), m.headline+m.viewHeight) - m.csrline - 1)
		m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline]) - nextCursorPosition
		return true
	}
	return readline.ENTER
}

type PrefixCommand struct {
	readline.KeyMap
}

func (*PrefixCommand) String() string {
	return "PREFIX-COMMAND"
}

func (cx *PrefixCommand) Call(ctx context.Context, B *readline.Buffer) readline.Result {
	key, err := B.GetKey()
	if err != nil {
		return readline.CONTINUE
	}
	f, ok := cx.KeyMap.Lookup(keys.Code(key))
	if !ok {
		return readline.CONTINUE
	}
	return f.Call(ctx, B)
}

func (m *Editor) init() error {
	if m.viewWidth > 0 {
		return nil
	}
	var err error
	m.viewWidth, m.viewHeight, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	m.viewHeight-- // for status line

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
	m.LineEditor.BindKey(keys.CtrlN, ac(m.CmdNextLine))
	m.LineEditor.BindKey(keys.CtrlP, ac(m.up))
	m.LineEditor.BindKey(keys.CtrlUp, ac(m.prevHistory))
	m.LineEditor.BindKey(keys.CtrlY, ac(m.paste))
	m.LineEditor.BindKey(keys.Delete, ac(m.joinBelow))
	m.LineEditor.BindKey(keys.Down, ac(m.CmdNextLine))
	m.LineEditor.BindKey(keys.Left, ac(m.left))
	m.LineEditor.BindKey(keys.PageDown, ac(m.nextHistory))
	m.LineEditor.BindKey(keys.PageUp, ac(m.prevHistory))
	m.LineEditor.BindKey(keys.Right, ac(m.right))
	m.LineEditor.BindKey(keys.Up, ac(m.up))
	m.LineEditor.BindKey(keys.CtrlM, ac(m.NewLine))
	m.LineEditor.BindKey(keys.CtrlJ, ac(m.Submit))
	m.LineEditor.BindKey(keys.CtrlR, readline.SelfInserter(keys.CtrlR))
	m.LineEditor.BindKey(keys.CtrlS, readline.SelfInserter(keys.CtrlS))

	escape := &PrefixCommand{}
	m.LineEditor.BindKey(keys.Escape, escape)
	escape.BindKey("p", ac(m.prevHistory)) // M-p: previous
	escape.BindKey("n", ac(m.nextHistory)) // M-n: next

	m.LineEditor.Init()
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

	m.lines = []string{}
	m.csrline = 0
	m.fixView()
	m.LineEditor.Cursor = 0
	if m.defaults != nil && len(m.defaults) > 0 {
		m.lines = append(m.lines, m.defaults...)
		if m.moveEnd {
			m.csrline = len(m.lines) - 1
			m.fixView()
			m.printAfter(m.headline)
			m.LineEditor.Out.WriteByte('\r')
			m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline])
		} else {
			m.escA(m.printAfter(0))
		}
	}
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
			m.printAfter(m.csrline)
			m.LineEditor.Out.WriteByte('\n')
			m.LineEditor.Out.Flush()
			return nil, err
		}
		m.LineEditor.Out.Flush()
		if !m.after(line) {
			return m.lines, nil
		}
		m.LineEditor.Out.Flush()
	}
}
