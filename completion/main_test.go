package completion

import (
	"strings"
	"testing"
)

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLineToFields(t *testing.T) {
	cases := [][]string{
		[]string{`aaa bbb ccc`, `aaa`, `bbb`, `ccc`},
		[]string{`aaa "b b " ccc`, `aaa`, `b b `, `ccc`},
		[]string{`aaa&ccc`, `aaa`, `&`, `ccc`},
	}
	for _, case1 := range cases {
		expect := case1[1:]
		result := lineToFields(case1[0], `"`, "&")
		if !eq(expect, result) {
			t.Fatalf("expect `%s`, but `%s`",
				strings.Join(expect, "|"), strings.Join(result, "|"))
		}
	}
}
