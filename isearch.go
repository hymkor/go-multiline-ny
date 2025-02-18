package multiline

import (
	"context"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nyaosorg/go-readline-ny"
)

func caseInsensitiveStringContains(s, t string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(t))
}

const (
	ansiCursorOff = "\x1B[?25l"

	// On Windows 8.1, the cursor is not shown immediately
	// without SetConsoleCursorPosition by `ESC[u`
	ansiCursorOn = "\x1B[?25h\x1B[s\x1B[u"
)

// NewLineMarkForIncrementalSearch is the string used instead of "\n". This variable is not guaranteed to remain valid in the future.
var NewLineMarkForIncrementalSearch = "\u21B2 "

func (m *Editor) cmdISearchBackward(ctx context.Context, this *readline.Buffer) readline.Result {
	var searchBuf strings.Builder
	foundStr := ""
	searchStr := ""
	lastFoundPos := this.History.Len() - 1

	moveOriginalLine := m.GotoEndLine()

	defer func() {
		io.WriteString(this.Out, "\x1B[2K")
		moveOriginalLine()
		this.Out.Flush()
		this.RepaintLastLine()
	}()

	update := func() {
		for i := this.History.Len() - 1; ; i-- {
			if i < 0 {
				foundStr = ""
				break
			}
			line := this.History.At(i)
			if caseInsensitiveStringContains(line, searchStr) {
				foundStr = line
				lastFoundPos = i
				break
			}
		}
	}
	for {
		drawStr := fmt.Sprintf("(i-search)[%s]:%s", searchStr,
			strings.ReplaceAll(foundStr, "\n", NewLineMarkForIncrementalSearch))
		drawWidth := readline.WidthT(0)
		this.Out.Write([]byte{'\r'})
		for _, ch := range readline.StringToMoji(drawStr) {
			w1 := ch.Width()
			if drawWidth+w1 >= this.ViewWidth() {
				break
			}
			ch.PrintTo(this.Out)
			drawWidth += w1
		}
		io.WriteString(this.Out, "\x1B[K"+ansiCursorOn)
		key, err := this.GetKey()
		if err != nil {
			println(err.Error())
			return readline.CONTINUE
		}
		io.WriteString(this.Out, ansiCursorOff)

		switch key {
		case "\b", "\x7F":
			searchBuf.Reset()
			// chop last char
			var lastchar rune
			for i, c := range searchStr {
				if i > 0 {
					searchBuf.WriteRune(lastchar)
				}
				lastchar = c
			}
			searchStr = searchBuf.String()
			update()
		case "\r":
			m.after = func(string) bool {
				m.clearLines()
				m.lines = strings.Split(foundStr, "\n")
				m.csrline = len(m.lines) - 1
				m.adjustHeadline()
				lfCount := m.printAfter(m.headline)
				lfCount -= (m.csrline - m.headline)
				m.up(lfCount)
				m.LineEditor.Cursor = 9999
				return true
			}
			return readline.ENTER
		case "\x03", "\x07", "\x1B":
			return readline.CONTINUE
		case "\x12":
			for i := lastFoundPos - 1; ; i-- {
				if i < 0 {
					i = this.History.Len() - 1
				}
				if i == lastFoundPos {
					break
				}
				line := this.History.At(i)
				if caseInsensitiveStringContains(line, searchStr) && foundStr != line {
					foundStr = line
					lastFoundPos = i
					break
				}
			}
		case "\x13":
			for i := lastFoundPos + 1; ; i++ {
				if i >= this.History.Len() {
					break
				}
				if i == lastFoundPos {
					break
				}
				line := this.History.At(i)
				if caseInsensitiveStringContains(line, searchStr) && foundStr != line {
					foundStr = line
					lastFoundPos = i
					break
				}
			}
		default:
			charcode, _ := utf8.DecodeRuneInString(key)
			if unicode.IsControl(charcode) {
				break
			}
			searchBuf.WriteRune(charcode)
			searchStr = searchBuf.String()
			update()
		}
	}
}
