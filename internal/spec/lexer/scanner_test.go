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

func TestNextTokenKeywords(t *testing.T) {
	keywords := []struct {
		input string
		kind  TokenKind
	}{
		{"spec", SPEC},
		{"type", TYPE},
		{"func", FUNC},
		{"predicate", PREDICATE},
		{"property", PROPERTY},
		{"forall", FORALL},
		{"where", WHERE},
		{"let", LET},
		{"require", REQUIRE},
		{"assert", ASSERT},
		{"is", IS},
		{"ok", OK},
		{"error", ERROR},
		{"and", AND},
		{"or", OR},
		{"not", NOT},
		{"in", IN},
		{"true", TRUE},
		{"false", FALSE},
		{"bind", BIND},
		{"target", TARGET},
		{"package", PACKAGE},
		{"bool", BOOL},
		{"int", INT},
		{"uint", UINT},
		{"float", FLOAT},
		{"string", STRING},
		{"bytes", BYTES},
		{"list", LIST},
		{"set", SET},
		{"map", MAP},
		{"option", OPTION},
	}
	for _, kw := range keywords {
		t.Run(kw.input, func(t *testing.T) {
			// arrange
			l := New(kw.input, "test.assay")

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
		kind    TokenKind
	}{
		{"foo", "foo", IDENT},
		{"myVar", "myVar", IDENT},
		{"_hidden", "_hidden", IDENT},
		{"x1", "x1", IDENT},
		{"Spec", "Spec", IDENT},     // capitalized — not a keyword
		{"FORALL", "FORALL", IDENT}, // uppercase — not a keyword
		{"a_b_c", "a_b_c", IDENT},
		{"_", "_", UNDERSCORE}, // lone underscore is UNDERSCORE, not IDENT
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// arrange
			l := New(tt.input, "test.assay")

			// act
			tok := l.NextToken()

			// assert
			require.Equal(t, tt.kind, tok.Kind)
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
		{"simple", `"hello"`, "hello"},
		{"empty", `""`, ""},
		{"escape newline", `"a\nb"`, "a\nb"},
		{"escape tab", `"a\tb"`, "a\tb"},
		{"escape quote", `"a\"b"`, `a"b`},
		{"escape backslash", `"a\\b"`, `a\b`},
		{"multiple escapes", `"\\n\t"`, "\\n\t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			l := New(tt.input, "test.assay")

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
			l := New(tt.input, "test.assay")

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
	// arrange
	l := New(`"hello\x"`, "test.assay")

	// act
	tok := l.NextToken()

	// assert
	require.Equal(t, ILLEGAL, tok.Kind)
	require.Equal(t, 1, tok.Pos.Column)
}

func TestNextTokenNumericLiterals(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		kind    TokenKind
		literal string
	}{
		{"integer", "42", INT_LIT, "42"},
		{"zero", "0", INT_LIT, "0"},
		{"multi-digit", "12345", INT_LIT, "12345"},
		{"float", "3.14", FLOAT_LIT, "3.14"},
		{"float leading zero", "0.5", FLOAT_LIT, "0.5"},
		{"float trailing digits", "1.000", FLOAT_LIT, "1.000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			l := New(tt.input, "test.assay")

			// act
			tok := l.NextToken()

			// assert
			require.Equal(t, tt.kind, tok.Kind)
			require.Equal(t, tt.literal, tok.Literal)
		})
	}
}

func TestNextTokenNumberDotDisambiguation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			kind    TokenKind
			literal string
		}
	}{
		{
			"int then dotdot",
			"42..100",
			[]struct {
				kind    TokenKind
				literal string
			}{
				{INT_LIT, "42"},
				{DOTDOT, ".."},
				{INT_LIT, "100"},
			},
		},
		{
			"int then dot",
			"42.foo",
			[]struct {
				kind    TokenKind
				literal string
			}{
				{INT_LIT, "42"},
				{DOT, "."},
				{IDENT, "foo"},
			},
		},
		{
			"float",
			"42.0",
			[]struct {
				kind    TokenKind
				literal string
			}{
				{FLOAT_LIT, "42.0"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			l := New(tt.input, "test.assay")

			for _, exp := range tt.expected {
				// act
				tok := l.NextToken()

				// assert
				require.Equal(t, exp.kind, tok.Kind)
				require.Equal(t, exp.literal, tok.Literal)
			}
		})
	}
}

func TestNextTokenOperatorsAndDelimiters(t *testing.T) {
	// arrange
	input := "== != < <= > >= + - * / % = .. -> ( ) { } [ ] , : ."
	expected := []struct {
		kind    TokenKind
		literal string
	}{
		{EQ, "=="}, {NEQ, "!="}, {LT, "<"}, {LTE, "<="}, {GT, ">"}, {GTE, ">="},
		{PLUS, "+"}, {MINUS, "-"}, {STAR, "*"}, {SLASH, "/"}, {PERCENT, "%"}, {ASSIGN, "="},
		{DOTDOT, ".."}, {ARROW, "->"},
		{LPAREN, "("}, {RPAREN, ")"}, {LBRACE, "{"}, {RBRACE, "}"},
		{LBRACKET, "["}, {RBRACKET, "]"}, {COMMA, ","}, {COLON, ":"}, {DOT, "."},
	}
	l := New(input, "test.assay")

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

func TestNextTokenMixedSequence(t *testing.T) {
	// arrange
	input := `spec MyTest {
    property "name" forall x in list {
      require x > 0
    }
  }`
	expected := []struct {
		kind    TokenKind
		literal string
	}{
		{SPEC, "spec"},
		{IDENT, "MyTest"},
		{LBRACE, "{"},
		{PROPERTY, "property"},
		{STRING_LIT, "name"},
		{FORALL, "forall"},
		{IDENT, "x"},
		{IN, "in"},
		{LIST, "list"},
		{LBRACE, "{"},
		{REQUIRE, "require"},
		{IDENT, "x"},
		{GT, ">"},
		{INT_LIT, "0"},
		{RBRACE, "}"},
		{RBRACE, "}"},
		{EOF, ""},
	}
	l := New(input, "test.assay")

	for _, exp := range expected {
		// act
		tok := l.NextToken()

		// assert
		require.Equal(t, exp.kind, tok.Kind, "expected %s (%q)", exp.kind, exp.literal)
		require.Equal(t, exp.literal, tok.Literal)
	}
}

func TestNextTokenPositionTracking(t *testing.T) {
	// arrange
	l := New("spec\n  foo", "test.assay")

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
	l := New("", "test.assay")

	// act
	tok := l.NextToken()

	// assert
	require.Equal(t, EOF, tok.Kind)
	require.Equal(t, "", tok.Literal)
}

func TestNextTokenIllegalCharacter(t *testing.T) {
	// arrange
	l := New("@", "test.assay")

	// act
	tok := l.NextToken()

	// assert
	require.Equal(t, ILLEGAL, tok.Kind)
	require.Equal(t, "@", tok.Literal)
	require.Equal(t, 1, tok.Pos.Column)
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
		t.Run(tt.word, func(t *testing.T) {
			// act
			got := LookupKeyword(tt.word)

			// assert
			require.Equal(t, tt.want, got)
		})
	}
}
