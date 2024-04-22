package multiline

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/nyaosorg/go-readline-ny/auto"
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
