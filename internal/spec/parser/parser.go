package parser

import (
	"fmt"

	"github.com/w-h-a/assay/internal/spec/ast"
	"github.com/w-h-a/assay/internal/spec/lexer"
)

// Error represents a parsing error with source position.
type Error struct {
	Message string
	Pos     ast.Position
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// Parse delegates lexing and then parses a token stream into a spec AST.
func Parse(source, file string) (*ast.SpecDecl, []Error) {
	tokens := lexer.Lex(source, file)
	p := &parser{tokens: tokens}
	spec := p.parseSpec()
	return spec, p.errors
}

// parser is a recursive-descent parser that walks a pre-lexed token slice.
// Errors are collected rather than halting, so a single parse pass can
// report multiple problems.
type parser struct {
	tokens []lexer.Token
	pos    int
	errors []Error
}

// parseSpec parses the top-level spec block: a name and a braced list
// of declarations.
func (p *parser) parseSpec() *ast.SpecDecl {
	start := p.peek()

	if _, ok := p.expect(lexer.SPEC); !ok {
		return &ast.SpecDecl{Pos: astPos(start)}
	}

	nameTok, ok := p.expect(lexer.STRING_LIT)
	if !ok {
		return &ast.SpecDecl{Pos: astPos(start)}
	}

	if _, ok := p.expect(lexer.LBRACE); !ok {
		return &ast.SpecDecl{Name: nameTok.Literal, Pos: astPos(start)}
	}

	var decls []ast.Decl
	for !p.at(lexer.RBRACE) && !p.at(lexer.EOF) {
		decl := p.parseDecl()
		if decl != nil {
			decls = append(decls, decl)
		}
	}

	p.expect(lexer.RBRACE)

	return &ast.SpecDecl{
		Name:         nameTok.Literal,
		Declarations: decls,
		Pos:          astPos(start),
	}
}

// parseDecl dispatches on the current token to the appropriate
// declaration parser.
func (p *parser) parseDecl() ast.Decl {
	switch p.peek().Kind {
	case lexer.TYPE:
		return p.parseTypeDecl()
	case lexer.FUNC:
		return p.parseFuncDecl()
	case lexer.PREDICATE:
		return p.parsePredicateDecl()
	default:
		tok := p.peek()
		p.addError(tok, "expected declaration, got %s", tok.Kind)
		p.skipToDecl()
		return nil
	}
}

// SPECIALIZED PARSERS

// parseTypeDecl parses a type declaration. A bare name like
// 'type Log' has no fields. A braced body like
// 'type Point { x: int, y: int }' has comma-separated fields.
func (p *parser) parseTypeDecl() *ast.TypeDecl {
	start := p.advance() // consume TYPE

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return &ast.TypeDecl{Pos: astPos(start)}
	}

	decl := &ast.TypeDecl{
		Name: nameTok.Literal,
		Pos:  astPos(start),
	}

	if !p.at(lexer.LBRACE) {
		return decl
	}

	p.advance() // consume LBRACE

	for !p.at(lexer.RBRACE) && !p.at(lexer.EOF) {
		before := p.pos
		field := p.parseFieldDecl()
		if p.pos == before {
			p.skipToField()
			if p.at(lexer.COMMA) {
				p.advance() // consume COMMA
			}
			continue
		}
		decl.Fields = append(decl.Fields, field)
		if p.at(lexer.COMMA) {
			p.advance()
		} else if !p.at(lexer.RBRACE) && !p.at(lexer.EOF) {
			p.addError(p.peek(), "expected comma between fields")
		}
	}

	p.expect(lexer.RBRACE)

	return decl
}

// parseFieldDecl parses a field name, colon, and type expression
// within a type declaration body.
func (p *parser) parseFieldDecl() ast.FieldDecl {
	start := p.peek()

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return ast.FieldDecl{Pos: astPos(start)}
	}

	if _, ok := p.expect(lexer.COLON); !ok {
		return ast.FieldDecl{Name: nameTok.Literal, Pos: astPos(nameTok)}
	}

	typ := p.parseTypeExpr()

	return ast.FieldDecl{
		Name: nameTok.Literal,
		Type: typ,
		Pos:  astPos(nameTok),
	}
}

// parseFuncDecl parses a function declaration:
//
//	func name(param1: Type1, param2: Type2) -> ReturnType
//	func name(params) -> (T1, T2)
//	func name(params)
func (p *parser) parseFuncDecl() *ast.FuncDecl {
	start := p.advance() // consume FUNC

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return &ast.FuncDecl{Pos: astPos(start)}
	}

	if _, ok := p.expect(lexer.LPAREN); !ok {
		return &ast.FuncDecl{Name: nameTok.Literal, Pos: astPos(start)}
	}

	var params []ast.Param
	for !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
		before := p.pos
		param := p.parseParam()
		if p.pos == before {
			p.skipToParam()
			if p.at(lexer.COMMA) {
				p.advance() // consume COMMA
			}
			continue
		}
		params = append(params, param)
		if p.at(lexer.COMMA) {
			p.advance() // consume COMMA
		} else if !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
			p.addError(p.peek(), "expected comma between parameters")
		}
	}

	p.expect(lexer.RPAREN)

	decl := &ast.FuncDecl{
		Name:   nameTok.Literal,
		Params: params,
		Pos:    astPos(start),
	}

	if !p.at(lexer.ARROW) {
		return decl
	}

	p.advance() // consume ARROW

	typ := p.parseTypeExpr()
	if len(typ.Elements) > 0 {
		decl.Returns = typ.Elements
	} else {
		decl.Returns = []ast.TypeExpr{typ}
	}

	return decl
}

// parsePredicateDecl parses a predicate declaration:
// predicate name(param1: Type1, param2: Type2) { expr }
func (p *parser) parsePredicateDecl() *ast.PredicateDecl {
	start := p.advance() // consume PREDICATE

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return &ast.PredicateDecl{Pos: astPos(start)}
	}

	if _, ok := p.expect(lexer.LPAREN); !ok {
		return &ast.PredicateDecl{Name: nameTok.Literal, Pos: astPos(start)}
	}

	var params []ast.Param
	for !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
		before := p.pos
		param := p.parseParam()
		if p.pos == before {
			p.skipToParam()
			if p.at(lexer.COMMA) {
				p.advance() // consume COMMA
			}
			continue
		}
		params = append(params, param)
		if p.at(lexer.COMMA) {
			p.advance() // consume COMMA
		} else if !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
			p.addError(p.peek(), "expected comma between parameters")
		}
	}

	p.expect(lexer.RPAREN)

	if _, ok := p.expect(lexer.LBRACE); !ok {
		return &ast.PredicateDecl{
			Name:   nameTok.Literal,
			Params: params,
			Pos:    astPos(start),
		}
	}

	body := p.parseExpr(precOr)

	p.expect(lexer.RBRACE)

	return &ast.PredicateDecl{
		Name:   nameTok.Literal,
		Params: params,
		Body:   body,
		Pos:    astPos(start),
	}
}

// parseParam parses a named, typed parameter: name: Type
func (p *parser) parseParam() ast.Param {
	start := p.peek()

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return ast.Param{Pos: astPos(start)}
	}

	if _, ok := p.expect(lexer.COLON); !ok {
		return ast.Param{Name: nameTok.Literal, Pos: astPos(nameTok)}
	}

	typ := p.parseTypeExpr()

	return ast.Param{
		Name: nameTok.Literal,
		Type: typ,
		Pos:  astPos(nameTok),
	}
}

// parseTypeExpr parses a type expression. A leading paren starts a
// tuple like (uint, error). Otherwise it expects a type name like int
// or Log, optionally followed by type parameters in brackets like
// list[int] or map[string, bytes].
func (p *parser) parseTypeExpr() ast.TypeExpr {
	if p.at(lexer.LPAREN) {
		return p.parseTupleType()
	}

	tok := p.peek()
	if !isTypeName(tok.Kind) {
		p.addError(tok, "expected type, got %s", tok.Kind)
		if !p.at(lexer.RPAREN) && !p.at(lexer.RBRACE) && !p.at(lexer.RBRACKET) && !p.at(lexer.COMMA) {
			p.advance()
		}
		return ast.TypeExpr{Pos: astPos(tok)}
	}
	p.advance() // consume type name

	te := ast.TypeExpr{
		Name: tok.Literal,
		Pos:  astPos(tok),
	}

	if p.at(lexer.LBRACKET) {
		p.advance() // consume LBRACKET
		te.Params = append(te.Params, p.parseTypeExpr())
		for p.at(lexer.COMMA) {
			p.advance() // consume COMMA
			if p.at(lexer.RBRACKET) {
				break
			}
			te.Params = append(te.Params, p.parseTypeExpr())
		}
		p.expect(lexer.RBRACKET)
	}

	return te
}

// parseTupleType parses parenthesized comma-separated list of types
// like (uint, error)
func (p *parser) parseTupleType() ast.TypeExpr {
	start := p.advance() // consume LPAREN

	te := ast.TypeExpr{Pos: astPos(start)}
	te.Elements = append(te.Elements, p.parseTypeExpr())
	for p.at(lexer.COMMA) {
		p.advance() // consume COMMA
		if p.at(lexer.RPAREN) {
			break
		}
		te.Elements = append(te.Elements, p.parseTypeExpr())
	}

	p.expect(lexer.RPAREN)

	return te
}

// parseExpr parses an expression using precedence climbing.
// minPrec is the minimum precedence for binary operators to bind.
// Recursing with prec+1 raises the bar so the inner call rejects
// operators at the same level, forcing them back to the outer call.
// This makes same-precedence operators left-associative.
func (p *parser) parseExpr(minPrec precedence) ast.Expr {
	left := p.parseAtom()

	for {
		prec := binaryPrec(p.peek().Kind)
		if prec == precNone || prec < minPrec {
			break
		}
		tok := p.advance() // consume operator
		right := p.parseExpr(prec + 1)
		left = &ast.BinaryExpr{
			Left:  left,
			Op:    tok.Literal,
			Right: right,
			Pos:   astPos(tok),
		}
	}

	return left
}

// parseAtom parses an atomic expression: literal, identifiers, or
// parenthesized expression. Unary prefix operators (-, not) are
// handled here as well.
func (p *parser) parseAtom() ast.Expr {
	tok := p.peek()

	switch tok.Kind {
	case lexer.MINUS, lexer.NOT:
		p.advance() // consume operator
		operand := p.parseAtom()
		return &ast.UnaryExpr{
			Op:      tok.Literal,
			Operand: operand,
			Pos:     astPos(tok),
		}
	case lexer.INT_LIT:
		p.advance()
		return &ast.LiteralExpr{Value: tok.Literal, Kind: ast.LiteralInt, Pos: astPos(tok)}
	case lexer.FLOAT_LIT:
		p.advance()
		return &ast.LiteralExpr{Value: tok.Literal, Kind: ast.LiteralFloat, Pos: astPos(tok)}
	case lexer.STRING_LIT:
		p.advance()
		return &ast.LiteralExpr{Value: tok.Literal, Kind: ast.LiteralString, Pos: astPos(tok)}
	case lexer.TRUE, lexer.FALSE:
		p.advance()
		return &ast.LiteralExpr{Value: tok.Literal, Kind: ast.LiteralBool, Pos: astPos(tok)}
	case lexer.IDENT:
		p.advance()
		return &ast.IdentExpr{Name: tok.Literal, Pos: astPos(tok)}
	case lexer.LPAREN:
		p.advance()
		expr := p.parseExpr(precOr)
		p.expect(lexer.RPAREN)
		return expr
	default:
		p.addError(tok, "expected expression, got %s", tok.Kind)
		if !p.at(lexer.RPAREN) && !p.at(lexer.RBRACE) && !p.at(lexer.RBRACKET) && !p.at(lexer.COMMA) {
			p.advance()
		}
		return &ast.IdentExpr{Pos: astPos(tok)}
	}
}

// PRIVATE HELPERS

// skipToDecl advances past tokens until a declaration start, closing brace,
// or EOF is found. Used for error recovery.
func (p *parser) skipToDecl() {
	for {
		switch p.peek().Kind {
		case lexer.TYPE, lexer.FUNC, lexer.PREDICATE, lexer.RBRACE, lexer.EOF:
			return
		}
		p.advance()
	}
}

// skipToField advances past tokens until a comma, closing brace,
// or EOF is found. Used for error recovery.
func (p *parser) skipToField() {
	for {
		switch p.peek().Kind {
		case lexer.COMMA, lexer.RBRACE, lexer.EOF:
			return
		}
		p.advance()
	}
}

// skipToParam advances past tokens until a comma, closing paren,
// or EOF is found. Used for error recovery.
func (p *parser) skipToParam() {
	for {
		switch p.peek().Kind {
		case lexer.COMMA, lexer.RPAREN, lexer.EOF:
			return
		}
		p.advance()
	}
}

func (p *parser) at(kind lexer.TokenKind) bool {
	return p.peek().Kind == kind
}

func (p *parser) expect(kind lexer.TokenKind) (lexer.Token, bool) {
	tok := p.peek()
	if tok.Kind == kind {
		return p.advance(), true
	}
	p.addError(tok, "expected %s, got %s", kind, tok.Kind)
	return tok, false
}

func (p *parser) advance() lexer.Token {
	tok := p.peek()
	if tok.Kind != lexer.EOF {
		p.pos++
	}
	return tok
}

func (p *parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) addError(tok lexer.Token, format string, args ...any) {
	p.errors = append(p.errors, Error{
		Message: fmt.Sprintf(format, args...),
		Pos:     astPos(tok),
	})
}
