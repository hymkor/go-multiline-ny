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
	LineEditor           readline.Editor
	Dirty                bool
	csrline              int
	lines                []string
	after                func(string) bool
	historyPtr           int
	viewWidth            int // when viewWidth==0, it means the instance is not initialized, yet
	viewHeight           int
	headline             int // the first line on the screen
	prompt               func(w io.Writer, i int) (int, error)
	defaults             []string
	moveEnd              bool
	modifiedHistoryEntry map[int]string
	promptLastLineOnly   bool
	StatusLineHeight     int
	Highlight            []readline.Highlight
	ResetColor           string
	DefaultColor         string

	memoHighlightSource string
	memoHighlightResult *readline.HighlightColorSequence
}

func (m *Editor) SetHistoryCycling(value bool)                  { m.LineEditor.HistoryCycling = value }
func (m *Editor) SetHistory(h readline.IHistory)                { m.LineEditor.History = h }
func (m *Editor) SetPrompt(f func(io.Writer, int) (int, error)) { m.prompt = f }
func (m *Editor) SetWriter(w io.Writer)                         { m.LineEditor.Writer = w }
func (m *Editor) SetDefault(d []string)                         { m.defaults = d }
func (m *Editor) SetMoveEnd(value bool)                         { m.moveEnd = value }
func (m *Editor) CursorLine() int                               { return m.csrline }
func (m *Editor) Lines() []string                               { return m.lines }

// Deprecated: set LineEditor.Highlight instead
func (m *Editor) SetColoring(c interface{}) {}

// SubmitOnEnterWhen defines the condition to submit when Enter-key is pressed.
func (m *Editor) SubmitOnEnterWhen(f func([]string, int) bool) {
	if f == nil {
		m.BindKey(keys.CtrlM, readline.AnonymousCommand(m.Submit))
		return
	}
	m.BindKey(keys.CtrlM, &readline.GoCommand{
		Name: "EnterToCommitWhen",
		Func: func(ctx context.Context, B *readline.Buffer) readline.Result {
			m.Sync(B.String())
			if f(m.lines, m.csrline) {
				return m.Submit(ctx, B)
			}
			return m.NewLine(ctx, B)
		},
	})
}

// Deprecated:
func (m *Editor) SwapEnter() error {
	m.BindKey(keys.CtrlM, readline.AnonymousCommand(m.Submit))
	m.BindKey(keys.CtrlJ, readline.AnonymousCommand(m.NewLine))
	return nil
}

func (m *Editor) Sync(line string) {
	if m.csrline >= len(m.lines) {
		m.lines = append(m.lines, line)
		m.Dirty = true
	} else {
		if m.lines[m.csrline] != line {
			m.Dirty = true
			m.lines[m.csrline] = line
		}
	}
}

func (m *Editor) up(n int) {
	if n == 0 {
		m.LineEditor.Out.Write([]byte{'\r'})
	} else if n > 0 {
		fmt.Fprintf(m.LineEditor.Out, "\x1B[%dF", n)
	}
}

func (m *Editor) CmdPreviousLine(ctx context.Context, rl *readline.Buffer) readline.Result {
	if m.csrline <= 0 {
		return m.CmdPreviousHistory(ctx, rl)
	}
	m.after = func(line string) bool {
		m.Sync(line)
		m.csrline--
		if m.fixView() < 0 {
			m.up(m.printAfter(m.csrline))
		} else {
			fmt.Fprint(m.LineEditor.Out, "\x1B[1F")
		}
		return true
	}
	return readline.ENTER
}

func (m *Editor) GotoEndLine() func() {
	end := min(len(m.lines), m.headline+m.viewHeight)
	if end < 1 {
		end = 1
	}
	lfCount := 0
	for i := m.csrline; i < end; i++ {
		fmt.Fprintln(m.LineEditor.Out)
		lfCount++
	}
	m.LineEditor.Out.Flush()
	return func() {
		if len(m.lines) >= m.viewHeight {
			io.WriteString(m.LineEditor.Out, "\x1B[1;1H")
			m.printAfter(m.headline)
			if lfCount > 1 {
				fmt.Fprintf(m.LineEditor.Out, "\x1B[%dF", lfCount-1)
			}
		} else if lfCount > 0 {
			fmt.Fprintf(m.LineEditor.Out, "\x1B[%dF", lfCount)
		}
		m.LineEditor.Out.Flush()
	}
}

func (m *Editor) Submit(_ context.Context, B *readline.Buffer) readline.Result {
	m.after = func(line string) bool {
		m.Sync(line)
		m.GotoEndLine()
		return false
	}
	return readline.ENTER
}

func (m *Editor) CmdNextLine(ctx context.Context, rl *readline.Buffer) readline.Result {
	if m.csrline >= len(m.lines)-1 {
		if m.LineEditor.History == nil || m.LineEditor.History.Len() <= 0 {
			return readline.CONTINUE
		}
		if m.historyPtr >= m.LineEditor.History.Len() {
			return readline.CONTINUE
		}
		m.CmdNextHistory(ctx, rl)
		m.after = m.printCurrentHistoryRecordAndGoToTop
		return readline.ENTER
	}
	m.after = func(line string) bool {
		m.Sync(line)
		m.csrline++
		if m.fixView() > 0 {
			m.up(m.csrline - m.headline)
			m.printAfter(m.headline)
		} else {
			fmt.Fprintln(m.LineEditor.Out)
		}
		return true
	}
	return readline.ENTER
}

func (m *Editor) CmdBackwardChar(ctx context.Context, b *readline.Buffer) readline.Result {
	if b.Cursor > 0 {
		return readline.CmdBackwardChar.Call(ctx, b)
	}
	if m.csrline == 0 {
		return readline.CONTINUE
	}
	m.after = func(line string) bool {
		m.Sync(line)
		m.csrline--
		if m.fixView() < 0 {
			m.up(m.printAfter(m.csrline))
		} else {
			fmt.Fprint(m.LineEditor.Out, "\x1B[1F")
		}
		m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline])
		return true
	}
	return readline.ENTER
}

func (m *Editor) CmdForwardChar(ctx context.Context, b *readline.Buffer) readline.Result {
	if b.Cursor < len(b.Buffer) {
		return readline.CmdForwardChar.Call(ctx, b)
	}
	if m.csrline+1 >= len(m.lines) {
		// To complete with the string of prediction
		return readline.CmdForwardChar.Call(ctx, b)
	}
	m.after = func(line string) bool {
		m.Sync(line)
		m.csrline++
		m.LineEditor.Cursor = 0
		if m.fixView() > 0 {
			m.up(m.csrline - m.headline)
			m.printAfter(m.headline)
		} else {
			fmt.Fprint(m.LineEditor.Out, "\n")
		}
		return true
	}
	return readline.ENTER
}

func (m *Editor) CmdBackwardDeleteChar(ctx context.Context, b *readline.Buffer) readline.Result {
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
			m.Dirty = true
			if m.csrline+1 < len(m.lines) {
				m.lines = deleteSliceAt(m.lines, m.csrline+1)
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
	m.Dirty = true

	b.RepaintAll()

	m.after = func(line string) bool {
		io.WriteString(m.LineEditor.Out, "\x1B[K\n")
		m.Sync(line)
		m.LineEditor.Cursor = 0
		m.csrline++
		m.fixView()
		m.up(m.printAfter(m.csrline))
		return true
	}
	return readline.ENTER
}

func (m *Editor) CmdDeleteChar(ctx context.Context, b *readline.Buffer) readline.Result {
	m.Dirty = true
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
		b.RepaintAfterPrompt()
		m.lines = deleteSliceAt(m.lines, m.csrline+1)
		io.WriteString(b.Out, "\x1B[s")
		fmt.Fprintln(m.LineEditor.Out)
		m.Sync(b.String())
		m.printAfter(m.csrline + 1)
		io.WriteString(b.Out, "\x1B[u")
		b.Out.Flush()
	}
	return readline.CONTINUE
}

func deleteSliceAt(array []string, at int) []string {
	copy(array[at:], array[at+1:])
	return array[:len(array)-1]
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

func printLastLine(p string, w io.Writer) {
	for {
		var line string
		var ok bool

		line, p, ok = strings.Cut(p, "\n")
		if !ok {
			io.WriteString(w, line)
			return
		}
		escapeStart := -1
		for i := 0; i < len(line); i++ {
			c := line[i]
			if c == '\x1B' {
				escapeStart = i
			}
			if escapeStart >= 0 && (('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')) {
				io.WriteString(w, line[escapeStart:i+1])
				escapeStart = -1
			}
		}
		if escapeStart >= 0 {
			io.WriteString(w, line[escapeStart:])
		}
	}
}

func (m *Editor) newPrinter() func(i int) {
	var colSeq *readline.HighlightColorSequence
	src := strings.Join(m.lines, "\n")

	if src == m.memoHighlightSource && m.memoHighlightResult != nil {
		colSeq = m.memoHighlightResult
	} else {
		colSeq = readline.HighlightToColoring(
			src,
			m.ResetColor,
			m.DefaultColor,
			m.Highlight)
		m.memoHighlightSource = src
		m.memoHighlightResult = colSeq
	}

	type LineColor struct {
		maps  []readline.EscapeSequenceId
		start readline.EscapeSequenceId
	}
	lineColors := []LineColor{}
	color := readline.NewEscapeSequenceId(m.ResetColor)
	colorMap := colSeq.ColorMap

	for i := 0; i < len(m.lines); i++ {
		lineColor1 := LineColor{
			maps:  colorMap[:len(m.lines[i])],
			start: color,
		}
		lineColors = append(lineColors, lineColor1)

		if len(m.lines[i]) >= len(colorMap) {
			break
		}
		color = colorMap[len(m.lines[i])]
		colorMap = colorMap[len(m.lines[i])+1:]
	}

	return func(i int) {
		var buffer strings.Builder
		m.prompt(&buffer, i)
		promptStr := buffer.String()
		if i != 0 || m.promptLastLineOnly {
			printLastLine(promptStr, m.LineEditor.Out)
		} else {
			io.WriteString(m.LineEditor.Out, promptStr)
		}
		m.promptLastLineOnly = true

		w0 := int(readline.GetStringWidth(cutEscapeSequenceAndOldLine(promptStr)))
		w := w0

		color := lineColors[i].start
		colorMap := lineColors[i].maps
		color.WriteTo(m.LineEditor.Out)

		for j, c := range m.lines[i] {
			newColor := colorMap[j]
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
		io.WriteString(m.LineEditor.Out, m.ResetColor)
		io.WriteString(m.LineEditor.Out, "\x1B[K")
	}
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
		printOne := m.newPrinter()
		for {
			printOne(i)
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
	m.up(lfCount)
	b.RepaintAll()
	return readline.CONTINUE
}

// fixView calculates the new value of m.headline
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

func (m *Editor) _printCurrentHistoryRecord(tail bool) {
	// clear
	m.up(m.csrline - m.headline)
	if value, ok := m.modifiedHistoryEntry[m.historyPtr]; ok {
		m.lines = strings.Split(value, "\n")
	} else if h := m.LineEditor.History; m.historyPtr < h.Len() {
		m.lines = strings.Split(h.At(m.historyPtr), "\n")
	}
	m.Dirty = true
	if tail {
		m.csrline = 0
	} else {
		m.csrline = len(m.lines) - 1
	}
	m.fixView()
	lfCount := m.printAfter(m.headline)
	lfCount -= (m.csrline - m.headline)
	m.up(lfCount)
	m.LineEditor.Cursor = 9999
}

func (m *Editor) printCurrentHistoryRecord(string) bool {
	m._printCurrentHistoryRecord(false)
	return true
}

func (m *Editor) printCurrentHistoryRecordAndGoToTop(string) bool {
	m._printCurrentHistoryRecord(true)
	return true
}

func (m *Editor) saveModfiedHistory(b *readline.Buffer) {
	m.Sync(b.String())
	alllines := strings.Join(m.lines, "\n")
	if m.historyPtr >= m.LineEditor.History.Len() ||
		m.LineEditor.History.At(m.historyPtr) != alllines {
		m.modifiedHistoryEntry[m.historyPtr] = alllines
	}
}

func (m *Editor) CmdPreviousHistory(_ context.Context, b *readline.Buffer) readline.Result {
	if m.LineEditor.History == nil || m.LineEditor.History.Len() <= 0 {
		return readline.CONTINUE
	}
	m.saveModfiedHistory(b)
	if m.historyPtr <= 0 {
		if !m.LineEditor.HistoryCycling {
			return readline.CONTINUE
		}
		m.historyPtr = m.LineEditor.History.Len()
	} else {
		m.historyPtr--
	}
	m.after = m.printCurrentHistoryRecord
	return readline.ENTER
}

func (m *Editor) CmdNextHistory(_ context.Context, b *readline.Buffer) readline.Result {
	if m.LineEditor.History == nil || m.LineEditor.History.Len() <= 0 {
		return readline.CONTINUE
	}
	m.saveModfiedHistory(b)
	if m.historyPtr >= m.LineEditor.History.Len() {
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

func (m *Editor) CmdYank(_ context.Context, b *readline.Buffer) readline.Result {
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
		m.Sync(line)
		fmt.Fprintln(m.LineEditor.Out)
		m.csrline++
		m.fixView()

		m.lines = insertSliceAt(m.lines, m.csrline, newlines)
		start := m.csrline
		m.csrline += len(newlines) - 1
		m.fixView()
		m.printAfter(start)
		m.up(min(len(m.lines), m.headline+m.viewHeight) - m.csrline - 1)
		m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline]) - nextCursorPosition
		return true
	}
	return readline.ENTER
}

type PrefixCommand struct {
	readline.KeyMap
	m      *Editor
	prompt string
}

func (m *Editor) NewPrefixCommand(prompt string) *PrefixCommand {
	return &PrefixCommand{
		prompt: prompt,
		m:      m,
	}
}

func (*PrefixCommand) String() string {
	return "PREFIX-COMMAND"
}

func (cx *PrefixCommand) Call(ctx context.Context, B *readline.Buffer) readline.Result {
	rewind := cx.m.GotoEndLine()
	io.WriteString(B.Out, cx.prompt)
	io.WriteString(B.Out, "\x1B[?25h")
	key, err := B.GetKey()
	fmt.Fprint(B.Out, "\x1B[?25l\x1B[2K")
	rewind()
	B.RepaintLastLine()
	if err != nil {
		return readline.CONTINUE
	}
	f, ok := cx.KeyMap.Lookup(keys.Code(key))
	if !ok {
		return readline.CONTINUE
	}
	return f.Call(ctx, B)
}

func (m *Editor) predictor(B *readline.Buffer) string {
	if m.csrline < len(m.lines)-1 {
		return ""
	}
	current := strings.TrimSpace(B.String())
	if len(current) <= 0 {
		return ""
	}
	for i := B.History.Len() - 1; i >= 0; i-- {
		lines := B.History.At(i)
		for _, line := range strings.Split(lines, "\n") {
			_line := strings.TrimSpace(line)
			if strings.HasPrefix(_line, current) {
				return _line[len(current):]
			}
		}
	}
	return ""
}

// SetPredictColor enables the prediction of go-readline-ny v1.5.0 and specify the colors
// (e.g.) `m.SetPredictColor([...]string{"\x1B[3;22;34m", "\x1B[23;39m"})`
func (m *Editor) SetPredictColor(colors [2]string) {
	m.LineEditor.Predictor = m.predictor
	m.LineEditor.PredictColor = colors
}

func (m *Editor) init() error {
	if m.modifiedHistoryEntry == nil {
		m.modifiedHistoryEntry = make(map[int]string)
	} else {
		for key := range m.modifiedHistoryEntry {
			delete(m.modifiedHistoryEntry, key)
		}
	}
	if m.viewWidth > 0 {
		return nil
	}
	var err error
	m.viewWidth, m.viewHeight, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	m.viewHeight -= m.StatusLineHeight

	m.LineEditor.LineFeedWriter = func(rc readline.Result, w io.Writer) (int, error) {
		if rc != readline.ENTER {
			return fmt.Fprintln(w)
		}
		return 0, nil
	}
	if m.prompt == nil {
		m.prompt = func(w io.Writer, i int) (int, error) {
			return fmt.Fprintf(w, "%2d ", i+1)
		}
	}
	m.LineEditor.PromptWriter = func(w io.Writer) (int, error) {
		if m.csrline != 0 || m.promptLastLineOnly {
			var buffer strings.Builder
			m.prompt(&buffer, m.csrline)
			promptStr := buffer.String()
			printLastLine(promptStr, w)
			return len(promptStr), nil
		}
		m.promptLastLineOnly = true
		return m.prompt(w, m.csrline)
	}

	type ac = readline.AnonymousCommand

	m.LineEditor.BindKey(keys.AltN, ac(m.CmdNextHistory))
	m.LineEditor.BindKey(keys.AltP, ac(m.CmdPreviousHistory))
	m.LineEditor.BindKey(keys.CtrlB, ac(m.CmdBackwardChar))
	m.LineEditor.BindKey(keys.CtrlD, ac(m.CmdDeleteChar))
	m.LineEditor.BindKey(keys.CtrlDown, ac(m.CmdNextHistory))
	m.LineEditor.BindKey(keys.CtrlF, ac(m.CmdForwardChar))
	m.LineEditor.BindKey(keys.CtrlH, ac(m.CmdBackwardDeleteChar))
	m.LineEditor.BindKey(keys.Backspace, ac(m.CmdBackwardDeleteChar))
	m.LineEditor.BindKey(keys.CtrlL, ac(m.repaint))
	m.LineEditor.BindKey(keys.CtrlN, ac(m.CmdNextLine))
	m.LineEditor.BindKey(keys.CtrlP, ac(m.CmdPreviousLine))
	m.LineEditor.BindKey(keys.CtrlUp, ac(m.CmdPreviousHistory))
	m.LineEditor.BindKey(keys.CtrlY, ac(m.CmdYank))
	m.LineEditor.BindKey(keys.Delete, ac(m.CmdDeleteChar))
	m.LineEditor.BindKey(keys.Down, ac(m.CmdNextLine))
	m.LineEditor.BindKey(keys.Left, ac(m.CmdBackwardChar))
	m.LineEditor.BindKey(keys.PageDown, ac(m.CmdNextHistory))
	m.LineEditor.BindKey(keys.PageUp, ac(m.CmdPreviousHistory))
	m.LineEditor.BindKey(keys.Right, ac(m.CmdForwardChar))
	m.LineEditor.BindKey(keys.Up, ac(m.CmdPreviousLine))
	m.LineEditor.BindKey(keys.CtrlM, ac(m.NewLine))
	m.LineEditor.BindKey(keys.CtrlJ, ac(m.Submit))
	m.LineEditor.BindKey(keys.CtrlR, ac(m.cmdISearchBackward))
	m.LineEditor.BindKey(keys.CtrlS, readline.SelfInserter(keys.CtrlS))

	escape := m.NewPrefixCommand("Esc-   [Enter] Submit, [Esc] Cancel, [p] Previous, [n] Next\rEsc-")
	m.LineEditor.BindKey(keys.Escape, escape)
	escape.BindKey("p", ac(m.CmdPreviousHistory)) // M-p: previous
	escape.BindKey("n", ac(m.CmdNextHistory))     // M-n: next
	escape.BindKey("\r", ac(m.Submit))            // M-Enter: submit

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

type fixPattern struct {
	Original interface{ FindAllStringIndex(string, int) [][]int }
	Prefix   string
	Postfix  string
}

func (f *fixPattern) FindAllStringIndex(s string, n int) [][]int {
	all := f.Prefix + "\n" + s + "\n" + f.Postfix
	result := [][]int{}
	for _, r := range f.Original.FindAllStringIndex(all, n) {
		start := r[0] - len(f.Prefix) - 1
		end := r[1] - len(f.Prefix) - 1
		if end < 0 {
			continue
		}
		if start >= len(s) {
			continue
		}
		if start < 0 {
			start = 0
		}
		if end >= len(s) {
			end = len(s)
		}
		if start < end {
			result = append(result, []int{start, end})
		}
	}
	return result
}

func (m *Editor) Read(ctx context.Context) ([]string, error) {
	if err := m.init(); err != nil {
		return nil, err
	}
	defer func() {
		m.promptLastLineOnly = false
	}()

	m.LineEditor.ResetColor = m.ResetColor
	m.LineEditor.DefaultColor = m.DefaultColor

	m.lines = []string{}
	m.csrline = 0
	m.fixView()
	m.LineEditor.Cursor = 0
	if len(m.defaults) > 0 {
		m.lines = append(m.lines, m.defaults...)
		if m.moveEnd {
			m.csrline = len(m.lines) - 1
			m.fixView()
			m.printAfter(m.headline)
			m.LineEditor.Out.WriteByte('\r')
			m.LineEditor.Cursor = readline.MojiCountInString(m.lines[m.csrline])
		} else {
			m.up(m.printAfter(0))
		}
	}
	if m.LineEditor.History != nil {
		m.historyPtr = m.LineEditor.History.Len()
	}
	if len(m.Highlight) > 0 {
		save := m.LineEditor.AfterCommand
		// Repaint after each typing
		m.LineEditor.AfterCommand = func(B *readline.Buffer) {
			m.Sync(B.String())
			m.up(m.csrline - m.headline)
			lfCount := m.printAfter(m.headline)
			m.up(lfCount - (m.csrline - m.headline))
			B.RepaintLastLine()
			if save != nil {
				save(B)
			}
		}
		defer func() {
			m.LineEditor.AfterCommand = save
		}()
	}
	for {
		if m.csrline < len(m.lines) {
			m.LineEditor.Default = m.lines[m.csrline]
		} else {
			m.LineEditor.Default = ""
		}
		m.after = func(string) bool { return true }
		if len(m.Highlight) > 0 {
			prefix := strings.Join(m.lines[:m.csrline], "\n") + "\n"
			postfix := ""
			if m.csrline+1 < len(m.lines) {
				postfix = "\n" + strings.Join(m.lines[m.csrline+1:], "\n")
			}
			newHighlight := make([]readline.Highlight, 0, len(m.Highlight))
			for _, h := range m.Highlight {
				newPattern := &fixPattern{
					Original: h.Pattern,
					Prefix:   prefix,
					Postfix:  postfix,
				}
				newHighlight = append(newHighlight,
					readline.Highlight{Pattern: newPattern, Sequence: h.Sequence})
			}
			m.LineEditor.Highlight = newHighlight
		}
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
