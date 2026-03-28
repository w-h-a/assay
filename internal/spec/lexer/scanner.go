package lexer

// Lexer performs lexical analysis on .assay source text.
type Lexer struct {
	source string
	file   string
	pos    int
	line   int
	column int
}

// New creates a Lexer for the given source text and file name.
func New(source, file string) *Lexer {
	return &Lexer{
		source: source,
		file:   file,
		line:   1,
		column: 1,
	}
}

// NextToken scans and returns the next token from the source.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	pos := l.Position()

	if l.isAtEnd() {
		return Token{Kind: EOF, Literal: "", Pos: pos}
	}

	ch := l.advance()

	switch ch {
	case '(':
		return Token{Kind: LPAREN, Literal: "(", Pos: pos}
	case ')':
		return Token{Kind: RPAREN, Literal: ")", Pos: pos}
	case '{':
		return Token{Kind: LBRACE, Literal: "{", Pos: pos}
	case '}':
		return Token{Kind: RBRACE, Literal: "}", Pos: pos}
	case '[':
		return Token{Kind: LBRACKET, Literal: "[", Pos: pos}
	case ']':
		return Token{Kind: RBRACKET, Literal: "]", Pos: pos}
	case ',':
		return Token{Kind: COMMA, Literal: ",", Pos: pos}
	case ':':
		return Token{Kind: COLON, Literal: ":", Pos: pos}
	case '+':
		return Token{Kind: PLUS, Literal: "+", Pos: pos}
	case '*':
		return Token{Kind: STAR, Literal: "*", Pos: pos}
	case '/':
		return Token{Kind: SLASH, Literal: "/", Pos: pos}
	case '=':
		if l.peek() == '=' {
			l.advance()
			return Token{Kind: EQ, Literal: "==", Pos: pos}
		}
		return Token{Kind: ASSIGN, Literal: "=", Pos: pos}
	case '!':
		if l.peek() == '=' {
			l.advance()
			return Token{Kind: NEQ, Literal: "!=", Pos: pos}
		}
		return Token{Kind: ILLEGAL, Literal: "!", Pos: pos}
	case '<':
		if l.peek() == '=' {
			l.advance()
			return Token{Kind: LTE, Literal: "<=", Pos: pos}
		}
		return Token{Kind: LT, Literal: "<", Pos: pos}
	case '>':
		if l.peek() == '=' {
			l.advance()
			return Token{Kind: GTE, Literal: ">=", Pos: pos}
		}
		return Token{Kind: GT, Literal: ">", Pos: pos}
	case '-':
		if l.peek() == '>' {
			l.advance()
			return Token{Kind: ARROW, Literal: "->", Pos: pos}
		}
		return Token{Kind: MINUS, Literal: "-", Pos: pos}
	case '.':
		if l.peek() == '.' {
			l.advance()
			return Token{Kind: DOTDOT, Literal: "..", Pos: pos}
		}
		return Token{Kind: DOT, Literal: ".", Pos: pos}
	case '_':
		if isAlphanumeric(l.peek()) {
			return l.scanIdentifier(pos)
		}
		return Token{Kind: UNDERSCORE, Literal: "_", Pos: pos}
	case '"':
		return l.scanString(pos)
	default:
		if isLetter(ch) {
			return l.scanIdentifier(pos)
		}
		if isDigit(ch) {
			return l.scanInteger(pos)
		}
		return Token{Kind: ILLEGAL, Literal: string(ch), Pos: pos}
	}
}

// Position returns the current source position (the location of the
// next character to be consumed).
func (l *Lexer) Position() Position {
	return Position{
		File:   l.file,
		Line:   l.line,
		Column: l.column,
	}
}

// scanIdentifier reads the rest of an identifier or keyword.
// The first character (letter or underscore) has already been consumed.
func (l *Lexer) scanIdentifier(pos Position) Token {
	start := l.pos - 1
	for isAlphanumeric(l.peek()) {
		l.advance()
	}
	literal := l.source[start:l.pos]
	return Token{Kind: LookupKeyword(literal), Literal: literal, Pos: pos}
}

// scanInteger reads the rest of an integer literal.
// The first digit has already been consumed.
func (l *Lexer) scanInteger(pos Position) Token {
	start := l.pos - 1
	for isDigit(l.peek()) {
		l.advance()
	}
	return Token{Kind: INT_LIT, Literal: l.source[start:l.pos], Pos: pos}
}

// scanString reads a double-quoted string literal.
// The opening quote has already been consumed.
func (l *Lexer) scanString(pos Position) Token {
	start := l.pos - 1
	var buf []byte
	for !l.isAtEnd() {
		ch := l.advance()
		switch ch {
		case '"': // end quote case
			return Token{Kind: STRING_LIT, Literal: string(buf), Pos: pos}
		case '\\': // single backslash case
			if l.isAtEnd() {
				return Token{Kind: ILLEGAL, Literal: l.source[start:l.pos], Pos: pos}
			}
			esc := l.advance()
			switch esc {
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case '"':
				buf = append(buf, '"')
			case '\\':
				buf = append(buf, '\\')
			default:
				return Token{Kind: ILLEGAL, Literal: l.source[start:l.pos], Pos: pos}
			}
		case '\n': // illegal new line case
			return Token{Kind: ILLEGAL, Literal: l.source[start : l.pos-1], Pos: pos}
		default:
			buf = append(buf, ch)
		}
	}
	return Token{Kind: ILLEGAL, Literal: l.source[start:l.pos], Pos: pos}
}

// peek returns the current character without consuming it.
// Returns 0 at end of source.
func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.pos]
}

// advance consumes the current character and returns it,
// updating line and column tracking.
func (l *Lexer) advance() byte {
	if l.isAtEnd() {
		return 0
	}
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return ch
}

// skipWhitespace advances past spaces, tabs, carriage returns, and newlines.
func (l *Lexer) skipWhitespace() {
	for !l.isAtEnd() {
		switch l.peek() {
		case ' ', '\t', '\r', '\n':
			l.advance()
		default:
			return
		}
	}
}

// isAtEnd reports whether all source text has been consumed.
func (l *Lexer) isAtEnd() bool {
	return l.pos >= len(l.source)
}
