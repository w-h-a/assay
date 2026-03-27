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

// Position returns the current source position (the location of the
// next character to be consumed).
func (l *Lexer) Position() Position {
	return Position{
		File:   l.file,
		Line:   l.line,
		Column: l.column,
	}
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

// current returns the most recently consumed character.
// Returns 0 if no characters have been consumed.
func (l *Lexer) current() byte {
	if l.pos == 0 {
		return 0
	}
	return l.source[l.pos-1]
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
