package lexer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLexer(t *testing.T) {
	// arrange
	source := "bind"
	file := "test.bind"

	// act
	l := New(source, file)

	// assert
	require.Equal(t, Position{File: "test.bind", Line: 1, Column: 1}, l.Position())
}

func TestNextTokenKeywords(t *testing.T) {
	keywords := []struct {
		input string
		kind  TokenKind
	}{
		{"bind", BIND},
		{"target", TARGET},
		{"package", PACKAGE},
		{"type", TYPE},
		{"func", FUNC},
	}
	for _, kw := range keywords {
		t.Run(kw.input, func(t *testing.T) {
			// arrange
			l := New(kw.input, "test.bind")

			// act
			tok := l.NextToken()

			// assert
			require.Equal(t, kw.kind, tok.Kind)
			require.Equal(t, kw.input, tok.Literal)
			require.Equal(t, 1, tok.Pos.Line)
			require.Equal(t, 1, tok.Pos.Column)

			// act — should be at end
			eof := l.NextToken()

			// assert
			require.Equal(t, EOF, eof.Kind)
		})
	}
}

func TestNextTokenIdentifiers(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"foo", "foo"},
		{"Log", "Log"},
		{"new_log", "new_log"},
		{"go", "go"},           // target value — not a keyword
		{"Bind", "Bind"},       // capitalized — not a keyword
		{"PACKAGE", "PACKAGE"}, // uppercase — not a keyword
		{"_private", "_private"},
		{"x1", "x1"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// arrange
			l := New(tt.input, "test.bind")

			// act
			tok := l.NextToken()

			// assert
			require.Equal(t, IDENT, tok.Kind)
			require.Equal(t, tt.literal, tok.Literal)
		})
	}
}

func TestNextTokenStringLiterals(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		literal string
	}{
		{"simple", `"log"`, "log"},
		{"empty", `""`, ""},
		{"package path", `"github.com/w-h-a/tally/internal/log"`, "github.com/w-h-a/tally/internal/log"},
		{"escape newline", `"a\nb"`, "a\nb"},
		{"escape tab", `"a\tb"`, "a\tb"},
		{"escape quote", `"a\"b"`, `a"b`},
		{"escape backslash", `"a\\b"`, `a\b`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			l := New(tt.input, "test.bind")

			// act
			tok := l.NextToken()

			// assert
			require.Equal(t, STRING_LIT, tok.Kind)
			require.Equal(t, tt.literal, tok.Literal)
			require.Equal(t, 1, tok.Pos.Column, "string starts at column 1")
		})
	}
}

func TestNextTokenUnterminatedString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"eof", `"hello`},
		{"newline", "\"hello\nworld\""},
		{"escape at eof", `"hello\`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			l := New(tt.input, "test.bind")

			// act
			tok := l.NextToken()

			// assert
			require.Equal(t, ILLEGAL, tok.Kind)
			require.Equal(t, 1, tok.Pos.Line)
			require.Equal(t, 1, tok.Pos.Column, "error position is the opening quote")
		})
	}
}

func TestNextTokenInvalidEscape(t *testing.T) {
	// arrange — a bad escape followed by trailing source. The scanner should
	// consume the entire string literal and leave the cursor positioned to
	// lex the trailing identifier cleanly (no cascade).
	l := New(`"hello\x world" foo`, "test.bind")

	// act
	tok := l.NextToken()

	// assert
	require.Equal(t, ILLEGAL, tok.Kind)
	require.Equal(t, `"hello\x world"`, tok.Literal, "ILLEGAL spans the entire malformed literal")
	require.Equal(t, 1, tok.Pos.Line)
	require.Equal(t, 1, tok.Pos.Column)

	// act — scanner should have parked just after the closing quote
	next := l.NextToken()

	// assert
	require.Equal(t, IDENT, next.Kind)
	require.Equal(t, "foo", next.Literal)
}

func TestNextTokenOperatorsAndDelimiters(t *testing.T) {
	// arrange
	input := "= { } ."
	expected := []struct {
		kind    TokenKind
		literal string
	}{
		{ASSIGN, "="},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{DOT, "."},
	}
	l := New(input, "test.bind")

	for _, exp := range expected {
		// act
		tok := l.NextToken()

		// assert
		require.Equal(t, exp.kind, tok.Kind)
		require.Equal(t, exp.literal, tok.Literal)
	}

	// act — should be at end
	eof := l.NextToken()

	// assert
	require.Equal(t, EOF, eof.Kind)
}

func TestNextTokenDottedPath(t *testing.T) {
	// arrange
	input := "log.Log"
	expected := []struct {
		kind    TokenKind
		literal string
	}{
		{IDENT, "log"},
		{DOT, "."},
		{IDENT, "Log"},
	}
	l := New(input, "test.bind")

	for _, exp := range expected {
		// act
		tok := l.NextToken()

		// assert
		require.Equal(t, exp.kind, tok.Kind)
		require.Equal(t, exp.literal, tok.Literal)
	}
}

func TestNextTokenPositionTracking(t *testing.T) {
	// arrange
	l := New("bind\n  foo", "test.bind")

	// act
	tok1 := l.NextToken()

	// assert
	require.Equal(t, 1, tok1.Pos.Line)
	require.Equal(t, 1, tok1.Pos.Column)

	// act — second token after newline + whitespace
	tok2 := l.NextToken()

	// assert
	require.Equal(t, 2, tok2.Pos.Line)
	require.Equal(t, 3, tok2.Pos.Column)
}

func TestNextTokenEOF(t *testing.T) {
	// arrange
	l := New("", "test.bind")

	// act
	tok := l.NextToken()

	// assert
	require.Equal(t, EOF, tok.Kind)
	require.Equal(t, "", tok.Literal)
}

func TestNextTokenIllegalCharacter(t *testing.T) {
	// arrange
	l := New("@", "test.bind")

	// act
	tok := l.NextToken()

	// assert
	require.Equal(t, ILLEGAL, tok.Kind)
	require.Equal(t, "@", tok.Literal)
	require.Equal(t, 1, tok.Pos.Column)
}

func TestLexComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			kind    TokenKind
			literal string
		}
	}{
		{
			"line comment skipped",
			"// this is a comment\nfoo",
			[]struct {
				kind    TokenKind
				literal string
			}{
				{IDENT, "foo"},
				{EOF, ""},
			},
		},
		{
			"inline comment",
			"foo // comment\nbar",
			[]struct {
				kind    TokenKind
				literal string
			}{
				{IDENT, "foo"},
				{IDENT, "bar"},
				{EOF, ""},
			},
		},
		{
			"comment at eof",
			"foo // end",
			[]struct {
				kind    TokenKind
				literal string
			}{
				{IDENT, "foo"},
				{EOF, ""},
			},
		},
		{
			"slash not followed by slash is illegal",
			"a / b",
			[]struct {
				kind    TokenKind
				literal string
			}{
				{IDENT, "a"},
				{ILLEGAL, "/"},
				{IDENT, "b"},
				{EOF, ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// act
			tokens := Lex(tt.input, "test.bind")

			// assert
			require.Len(t, tokens, len(tt.expected))
			for i, exp := range tt.expected {
				require.Equal(t, exp.kind, tokens[i].Kind, "token %d kind", i)
				require.Equal(t, exp.literal, tokens[i].Literal, "token %d literal", i)
			}
		})
	}
}

func TestLexContinuesAfterIllegal(t *testing.T) {
	// act
	tokens := Lex("@foo", "test.bind")

	// assert
	require.Len(t, tokens, 3)

	require.Equal(t, ILLEGAL, tokens[0].Kind)
	require.Equal(t, "@", tokens[0].Literal)
	require.Equal(t, 1, tokens[0].Pos.Line)
	require.Equal(t, 1, tokens[0].Pos.Column)

	require.Equal(t, IDENT, tokens[1].Kind)
	require.Equal(t, "foo", tokens[1].Literal)
	require.Equal(t, 1, tokens[1].Pos.Line)
	require.Equal(t, 2, tokens[1].Pos.Column)

	require.Equal(t, EOF, tokens[2].Kind)
}

func TestLexLogBind(t *testing.T) {
	// arrange
	input := `bind "log" {
    target go
    package "github.com/w-h-a/tally/internal/log"
    type Log     = log.Log
    func new_log = log.New
    func append  = Log.Append
    func read    = Log.ReadAt
  }`
	expected := []struct {
		kind    TokenKind
		literal string
	}{
		{BIND, "bind"},
		{STRING_LIT, "log"},
		{LBRACE, "{"},
		// target go
		{TARGET, "target"},
		{IDENT, "go"},
		// package "github.com/..."
		{PACKAGE, "package"},
		{STRING_LIT, "github.com/w-h-a/tally/internal/log"},
		// type Log = log.Log
		{TYPE, "type"},
		{IDENT, "Log"},
		{ASSIGN, "="},
		{IDENT, "log"},
		{DOT, "."},
		{IDENT, "Log"},
		// func new_log = log.New
		{FUNC, "func"},
		{IDENT, "new_log"},
		{ASSIGN, "="},
		{IDENT, "log"},
		{DOT, "."},
		{IDENT, "New"},
		// func append = Log.Append
		{FUNC, "func"},
		{IDENT, "append"},
		{ASSIGN, "="},
		{IDENT, "Log"},
		{DOT, "."},
		{IDENT, "Append"},
		// func read = Log.ReadAt
		{FUNC, "func"},
		{IDENT, "read"},
		{ASSIGN, "="},
		{IDENT, "Log"},
		{DOT, "."},
		{IDENT, "ReadAt"},
		{RBRACE, "}"},
		{EOF, ""},
	}

	// act
	tokens := Lex(input, "test.bind")

	// assert
	require.Len(t, tokens, len(expected))
	for i, exp := range expected {
		require.Equal(t, exp.kind, tokens[i].Kind, "token %d: expected %s, got %s", i, exp.kind, tokens[i].Kind)
		require.Equal(t, exp.literal, tokens[i].Literal, "token %d literal", i)
	}
}

func TestPositionString(t *testing.T) {
	tests := []struct {
		pos  Position
		want string
	}{
		{Position{File: "test.bind", Line: 1, Column: 5}, "test.bind:1:5"},
		{Position{File: "", Line: 3, Column: 10}, "3:10"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			// act
			got := tt.pos.String()

			// assert
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLookupKeyword(t *testing.T) {
	tests := []struct {
		word string
		want TokenKind
	}{
		{"bind", BIND},
		{"target", TARGET},
		{"package", PACKAGE},
		{"type", TYPE},
		{"func", FUNC},
		{"go", IDENT},      // target value, not a keyword
		{"Bind", IDENT},    // case-sensitive
		{"PACKAGE", IDENT}, // case-sensitive
		{"myVar", IDENT},
	}
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			// act
			got := LookupKeyword(tt.word)

			// assert
			require.Equal(t, tt.want, got)
		})
	}
}
