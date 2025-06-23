package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-colorable"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/keys"
	"github.com/nyaosorg/go-readline-ny/simplehistory"

	"github.com/hymkor/go-multiline-ny"
	"github.com/hymkor/go-multiline-ny/completion"
)

var (
	commands = []string{"select", "insert", "delete", "update"}
	tables   = []string{"dept", "emp", "bonus", "salgrade", "bonus"}
	columns  = []string{"deptno", "dname", "loc", "empno", "ename", "job", "mgr", "hiredate", "sal", "comm", "grade", "losal", "hisal"}
)

func getCompletionCandidates(fields []string) (forCompletion []string, forListing []string) {
	candidates := commands
	for _, word := range fields {
		if strings.EqualFold(word, "from") {
			candidates = append([]string{"where"}, tables...)
		} else if strings.EqualFold(word, "set") {
			candidates = append([]string{"where"}, columns...)
		} else if strings.EqualFold(word, "update") {
			candidates = append([]string{"set"}, tables...)
		} else if strings.EqualFold(word, "delete") {
			candidates = []string{"from"}
		} else if strings.EqualFold(word, "select") {
			candidates = append([]string{"from"}, columns...)
		} else if strings.EqualFold(word, "where") {
			candidates = append([]string{"and", "or"}, columns...)
		}
	}
	return candidates, candidates
}

func main() {
	ctx := context.Background()
	fmt.Println("C-m or Enter      : Insert a linefeed")
	fmt.Println("C-p or UP         : Move to the previous line.")
	fmt.Println("C-n or DOWN       : Move to the next line")
	fmt.Println("C-j or Esc+Enter  : Submit")
	fmt.Println("C-c               : Abort.")
	fmt.Println("C-D with no chars : Quit.")
	fmt.Println("C-UP   or M-P     : Move to the previous history entry")
	fmt.Println("C-DOWN or M-N     : Move to the next history entry")

	var ed multiline.Editor
	ed.SetPrompt(func(w io.Writer, lnum int) (int, error) {
		return fmt.Fprintf(w, "[%d] ", lnum+1)
	})
	ed.SetPredictColor(readline.PredictColorBlueItalic)

	ed.Highlight = []readline.Highlight{
		// Words -> dark green
		{Pattern: regexp.MustCompile(`(?i)(SELECT|INSERT|FROM|WHERE|AS)`), Sequence: "\x1B[33;49;22m"},
		// Double quotation -> light magenta
		{Pattern: regexp.MustCompile(`(?m)"([^"\n]*\\")*[^"\n]*$|"([^"\n]*\\")*[^"\n]*"`), Sequence: "\x1B[32;49;1m"},
		// Single quotation -> light red
		{Pattern: regexp.MustCompile(`(?m)'([^'\n]*\\')*[^'\n]*$|'([^'\n]*\\')*[^'\n]*'`), Sequence: "\x1B[31;49;1m"},
		// Number literal -> light blue
		{Pattern: regexp.MustCompile(`[0-9]+`), Sequence: "\x1B[34;49;1m"},
		// Comment -> dark gray
		{Pattern: regexp.MustCompile(`(?s)/\*.*?\*/`), Sequence: "\x1B[30;49;1m"},
		// Multiline string literal -> dark red
		{Pattern: regexp.MustCompile("(?s)```.*?```"), Sequence: "\x1B[31;49;22m"},
	}
	ed.ResetColor = "\x1B[0m"
	ed.DefaultColor = "\x1B[37;49;1m"

	// To enable escape sequence on Windows.
	// (On other operating systems, it can be ommited)
	ed.SetWriter(colorable.NewColorableStdout())

	// enable history (optional)
	history := simplehistory.New()
	ed.SetHistory(history)
	ed.SetHistoryCycling(true)

	// enable completion (optional)
	ed.BindKey(keys.CtrlI, &completion.CmdCompletionOrList{
		// Characters listed here are excluded from completion.
		Delimiter: "&|><;",
		// Enclose candidates with these characters when they contain spaces
		Enclosure: `"'`,
		// String to append when only one candidate remains
		Postfix: " ",
		// Function for listing candidates
		Candidates: getCompletionCandidates,
	})

	for {
		lines, err := ed.Read(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		L := strings.Join(lines, "\n")
		fmt.Println("-----")
		fmt.Println(L)
		fmt.Println("-----")
		history.Add(L)
	}
}
