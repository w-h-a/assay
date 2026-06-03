package is

func Alphanumeric(ch byte) bool {
	return Letter(ch) || Digit(ch) || ch == '_'
}

func Letter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func Digit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
