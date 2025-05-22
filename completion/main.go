package completion

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nyaosorg/go-box/v2"
	"github.com/nyaosorg/go-readline-ny"
	singleCompletion "github.com/nyaosorg/go-readline-ny/completion"

	"github.com/hymkor/go-multiline-ny"
)

type CmdCompletion = singleCompletion.CmdCompletion2

type CmdCompletionOrList struct {
	editor     *multiline.Editor
	Delimiter  string
	Enclosure  string
	Postfix    string
	Candidates func(fieldsBeforeCursor []string) (completionSet []string, listingSet []string)
}

func (C *CmdCompletionOrList) SetEditor(m *multiline.Editor) {
	C.editor = m
}

func (C *CmdCompletionOrList) String() string {
	return "MULTI_COMPLETION_OR_LIST"
}

func removeQuotes(s, q string) string {
	var buffer strings.Builder
	for _, c := range s {
		if !strings.ContainsRune(q, c) {
			buffer.WriteRune(c)
		}
	}
	return buffer.String()
}

func lineToFields(line, quotes, del string) (fields []string) {
	const spaces = " \t\r\n\v\f"
	for len(line) > 0 {
		for len(line) > 0 {
			c, siz := utf8.DecodeRuneInString(line)
			if !strings.ContainsRune(spaces, c) {
				break
			}
			line = line[siz:]
		}
		bits := 0
		i := 0
		for {
			if i >= len(line) {
				if len(line) > 0 {
					fields = append(fields, removeQuotes(line, quotes))
				}
				break
			}
			c, siz := utf8.DecodeRuneInString(line[i:])

			if j := strings.IndexRune(quotes, c); j >= 0 {
				bits ^= (1 << j)
			} else if bits == 0 {
				if strings.ContainsRune(spaces, c) {
					fields = append(fields, removeQuotes(line[:i], quotes))
					break
				}
				if strings.ContainsRune(del, c) {
					fields = append(fields, removeQuotes(line[:i], quotes))
					fields = append(fields, string(c))
					i += siz
					break
				}
			}
			i += siz
		}
		line = line[i:]
	}
	return
}

func (C *CmdCompletionOrList) Call(ctx context.Context, B *readline.Buffer) readline.Result {
	fieldsBeforeCurrentLine := []string{}
	for _, line := range C.editor.Lines()[:C.editor.CursorLine()] {
		f := lineToFields(line, C.Enclosure, C.Delimiter)
		fieldsBeforeCurrentLine = append(fieldsBeforeCurrentLine, f...)
	}

	newCandidates := func(fieldsBeforeCursor []string) ([]string, []string) {
		f := make([]string, 0, len(fieldsBeforeCurrentLine)+len(fieldsBeforeCursor))
		f = append(f, fieldsBeforeCurrentLine...)
		f = append(f, fieldsBeforeCursor...)
		return C.Candidates(f)
	}
	list := singleCompletion.Complete(C.Enclosure, C.Delimiter, B, newCandidates, C.Postfix)
	if len(list) <= 0 {
		return readline.CONTINUE
	}
	// listing
	m := C.editor
	m.SetNextEditHook(func(line string) bool {
		m.GotoEndLine()

		box.Print(ctx, list, B.Out)

		lfCount := m.PrintFromLine(m.Headline())
		lfCount -= (m.CursorLine() - m.Headline())
		if lfCount > 0 {
			fmt.Fprintf(B.Out, "\x1B[%dA", lfCount)
		}
		B.RepaintLastLine()
		return true
	})
	return readline.ENTER
}
