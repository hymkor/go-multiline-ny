package multiline

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/nyaosorg/go-ttyadapter/auto"
	"github.com/nyaosorg/go-readline-ny/keys"
)

func TestEditorRead(t *testing.T) {
	// Type `SELECT ALL\n  FROM\nDUAL`
	keyin := strings.Split("SELECT ALL\r  FROM\rDUAL", "")
	// Move cursor to the top line
	keyin = append(keyin, keys.Up, keys.Up)
	// Replace `ALL` to `*`
	keyin = append(keyin, keys.CtrlE, keys.CtrlH, keys.CtrlH, keys.CtrlH, "*", keys.CtrlJ)

	var ed Editor
	ed.LineEditor.Tty = &auto.Pilot{Text: keyin}
	ed.SetWriter(io.Discard)
	lines, err := ed.Read(context.Background())
	if err != nil {
		t.Fatal(err.Error())
	}
	result := strings.Join(lines, "\n")
	expect := "SELECT *\n  FROM\nDUAL"
	if result != expect {
		t.Fatalf("expect %#v, but %#v", expect, result)
	}
}

func TestPrintLastLine(t *testing.T) {
	cases := [][2]string{
		[2]string{"test\nprompt", "prompt"},
		[2]string{"test\x1B[32m\nprompt\x1B[0m", "\x1B[32mprompt\x1B[0m"},
	}
	for _, case1 := range cases {
		var buffer strings.Builder
		printLastLine(case1[0], &buffer)
		result := buffer.String()

		if case1[1] != result {
			t.Fatalf("expect %s, but %s", case1[1], result)
		}
	}
}
