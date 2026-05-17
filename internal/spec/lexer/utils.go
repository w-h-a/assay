package lexer

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlphanumeric(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '_'
}
