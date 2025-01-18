package multiline

import (
	"io"

	"github.com/nyaosorg/go-readline-ny"
)

// copy from go-readline-ny/highlight

type escapeSequenceId uint

var (
	escapeSequences          = []string{}
	escapeSequenceStringToId = map[string]escapeSequenceId{}
)

type colorInterface interface {
	io.WriterTo
	Equals(colorInterface) bool
}

func newEscapeSequenceId(s string) escapeSequenceId {
	if code, ok := escapeSequenceStringToId[s]; ok {
		return code
	}
	code := escapeSequenceId(len(escapeSequences))
	escapeSequences = append(escapeSequences, s)
	escapeSequenceStringToId[s] = code
	return code
}

func (e escapeSequenceId) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, escapeSequences[e])
	return int64(n), err
}

func (e escapeSequenceId) Equals(other colorInterface) bool {
	o, ok := other.(escapeSequenceId)
	return ok && o == e
}

type highlightColorSequence struct {
	colorMap []escapeSequenceId
	index    int
	resetSeq escapeSequenceId
}

func highlightToColoring(input string, resetColor, defaultColor string, H []readline.Highlight) *highlightColorSequence {
	colorMap := make([]escapeSequenceId, len(input))
	defaultSeq := newEscapeSequenceId(defaultColor)
	for i := 0; i < len(input); i++ {
		colorMap[i] = defaultSeq
	}
	for _, h := range H {
		positions := h.Pattern.FindAllStringIndex(input, -1)
		if positions == nil {
			continue
		}
		seq := newEscapeSequenceId(h.Sequence)
		for _, p := range positions {
			for i := p[0]; i < p[1]; i++ {
				colorMap[i] = seq
			}
		}
	}
	return &highlightColorSequence{
		colorMap: colorMap,
		resetSeq: newEscapeSequenceId(resetColor),
	}
}
