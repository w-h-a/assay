package lexer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLexer(t *testing.T) {
	// arrange
	source := "hello"
	file := "test.assay"

	// act
	l := New(source, file)

	// assert
	require.Equal(t, Position{File: "test.assay", Line: 1, Column: 1}, l.Position())
}

func TestPositionString(t *testing.T) {
	tests := []struct {
		pos  Position
		want string
	}{
		{Position{File: "test.assay", Line: 1, Column: 5}, "test.assay:1:5"},
		{Position{File: "", Line: 3, Column: 10}, "3:10"},
	}
	for _, tt := range tests {
		// act
		got := tt.pos.String()

		// assert
		require.Equal(t, tt.want, got)
	}
}

func TestAdvanceTracksPosition(t *testing.T) {
	// arrange
	l := New("ab\ncd\nef", "test.assay")

	steps := []struct {
		ch   byte
		line int
		col  int
	}{
		{'a', 1, 2},
		{'b', 1, 3},
		{'\n', 2, 1},
		{'c', 2, 2},
		{'d', 2, 3},
		{'\n', 3, 1},
		{'e', 3, 2},
		{'f', 3, 3},
	}

	for i, s := range steps {
		// act
		ch := l.advance()
		pos := l.Position()

		// assert
		require.Equal(t, s.ch, ch, "step %d: advance()", i)
		require.Equal(t, s.line, pos.Line, "step %d: line", i)
		require.Equal(t, s.col, pos.Column, "step %d: column", i)
	}

	require.True(t, l.isAtEnd())
}

func TestPeekDoesNotAdvance(t *testing.T) {
	// arrange
	l := New("ab", "test.assay")

	// act
	ch := l.peek()
	pos := l.Position()

	// assert
	require.Equal(t, byte('a'), ch)
	require.Equal(t, 1, pos.Line)
	require.Equal(t, 1, pos.Column)

	// act — peek again
	ch2 := l.peek()

	// assert — idempotent
	require.Equal(t, byte('a'), ch2)
}

func TestCurrent(t *testing.T) {
	// arrange
	l := New("ab", "test.assay")

	// act + assert — before any advance
	require.Equal(t, byte(0), l.current())

	// act + assert — after first advance
	l.advance()
	require.Equal(t, byte('a'), l.current())

	// act + assert — after second advance
	l.advance()
	require.Equal(t, byte('b'), l.current())
}

func TestSkipWhitespace(t *testing.T) {
	// arrange
	l := New("  \t\n  hello", "test.assay")

	// act
	l.skipWhitespace()

	// assert
	pos := l.Position()
	require.Equal(t, 2, pos.Line)
	require.Equal(t, 3, pos.Column)
	require.Equal(t, byte('h'), l.peek())
}

func TestSkipWhitespaceAtEnd(t *testing.T) {
	// arrange
	l := New("   ", "test.assay")

	// act
	l.skipWhitespace()

	// assert
	require.True(t, l.isAtEnd())
}

func TestAdvancePastEnd(t *testing.T) {
	// arrange
	l := New("a", "test.assay")
	l.advance()

	// act
	ch := l.advance()
	pk := l.peek()

	// assert
	require.True(t, l.isAtEnd())
	require.Equal(t, byte(0), ch)
	require.Equal(t, byte(0), pk)
}

func TestEmptySource(t *testing.T) {
	// arrange
	l := New("", "test.assay")

	// act
	pk := l.peek()
	ch := l.advance()

	// assert
	require.True(t, l.isAtEnd())
	require.Equal(t, byte(0), pk)
	require.Equal(t, byte(0), ch)
}

func TestLookupKeyword(t *testing.T) {
	tests := []struct {
		word string
		want TokenKind
	}{
		{"spec", SPEC},
		{"forall", FORALL},
		{"true", TRUE},
		{"bind", BIND},
		{"int", INT},
		{"option", OPTION},
		{"myVar", IDENT},
		{"Spec", IDENT},   // case-sensitive
		{"FORALL", IDENT}, // case-sensitive
	}
	for _, tt := range tests {
		// act
		got := LookupKeyword(tt.word)

		// assert
		require.Equal(t, tt.want, got, "LookupKeyword(%q)", tt.word)
	}
}
