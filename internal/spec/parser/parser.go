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
	case lexer.PROPERTY:
		return p.parsePropertyDecl()
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

// parsePropertyDecl parses a property declaration:
// property name forall(x: int, n: int in 1..100) { body }
func (p *parser) parsePropertyDecl() *ast.PropertyDecl {
	start := p.advance() // consume PROPERTY

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return &ast.PropertyDecl{Pos: astPos(start)}
	}

	if _, ok := p.expect(lexer.FORALL); !ok {
		return &ast.PropertyDecl{Name: nameTok.Literal, Pos: astPos(start)}
	}

	forall := p.parseForallClause()

	if _, ok := p.expect(lexer.LBRACE); !ok {
		return &ast.PropertyDecl{
			Name:   nameTok.Literal,
			Forall: forall,
			Pos:    astPos(start),
		}
	}

	body := p.parsePropertyBody()

	p.expect(lexer.RBRACE)

	shape := ast.Contractual
	for _, s := range body {
		switch s.(type) {
		case *ast.LetBinding, *ast.RequireStmt:
			shape = ast.Sequential
		}
		if shape == ast.Sequential {
			break
		}
	}

	return &ast.PropertyDecl{
		Name:   nameTok.Literal,
		Forall: forall,
		Body:   body,
		Shape:  shape,
	}
}

// parseForallClause parses the forall(var1, var2, ...) clause.
func (p *parser) parseForallClause() ast.ForallClause {
	start := p.peek()

	if _, ok := p.expect(lexer.LPAREN); !ok {
		return ast.ForallClause{Pos: astPos(start)}
	}

	var vars []ast.QuantifiedVar
	for !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
		before := p.pos
		v := p.parseQuantifiedVar()
		if p.pos == before {
			p.skipToParam()
			if p.at(lexer.COMMA) {
				p.advance() // consume COMMA
			}
			continue
		}
		vars = append(vars, v)
		if p.at(lexer.COMMA) {
			p.advance() // consume COMMA
		} else if !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
			p.addError(p.peek(), "expected comma between forall variables")
		}
	}

	p.expect(lexer.RPAREN)

	return ast.ForallClause{Vars: vars, Pos: astPos(start)}
}

// parseQuantifiedVar parses a quantified variable: name: Type or name: Type in <generator>
func (p *parser) parseQuantifiedVar() ast.QuantifiedVar {
	start := p.peek()

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return ast.QuantifiedVar{Pos: astPos(start)}
	}

	if _, ok := p.expect(lexer.COLON); !ok {
		return ast.QuantifiedVar{Name: nameTok.Literal, Pos: astPos(start)}
	}

	typ := p.parseTypeExpr()

	qv := ast.QuantifiedVar{
		Name: nameTok.Literal,
		Type: typ,
		Pos:  astPos(nameTok),
	}

	if !p.at(lexer.IN) {
		return qv
	}

	p.advance() // consume IN
	qv.Generator = p.parseGenerator()

	return qv
}

// parseGenerator parses a generator constraint after 'in':
// 1..100 -> RangeGen
// strings(1, 50) -> BuiltinGen
// one_of(a, b) -> OneOfGen
func (p *parser) parseGenerator() ast.GeneratorConstraint {
	// Range
	if (p.at(lexer.INT_LIT) || p.at(lexer.FLOAT_LIT) || p.at(lexer.IDENT)) && p.peekAt(1).Kind == lexer.DOTDOT {
		lo := p.parseAtom()
		pos := p.peek()
		p.advance() // consume DOTDOT
		hi := p.parseAtom()
		return &ast.RangeGen{Lo: lo, Hi: hi, Pos: astPos(pos)}
	}

	// BuiltinGen or OneOfGen
	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return nil
	}

	if _, ok := p.expect(lexer.LPAREN); !ok {
		return nil
	}

	var args []ast.Expr
	for !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
		before := p.pos
		arg := p.parseExpr(precOr)
		if p.pos == before {
			p.skipToParam()
			if p.at(lexer.COMMA) {
				p.advance() // consume COMMA
			}
			continue
		}
		args = append(args, arg)
		if p.at(lexer.COMMA) {
			p.advance() // consume COMMA
		} else if !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
			p.addError(p.peek(), "expected comma between arguments")
		}
	}

	p.expect(lexer.RPAREN)

	if nameTok.Literal == "one_of" {
		return &ast.OneOfGen{Values: args, Pos: astPos(nameTok)}
	}

	return &ast.BuiltinGen{Name: nameTok.Literal, Args: args, Pos: astPos(nameTok)}
}

// parsePropertyBody parses the statements inside a property body.
// The body is a sequence of let-bindings and require guards,
// ending with a terminal assertion expression.
func (p *parser) parsePropertyBody() []ast.Stmt {
	var stmts []ast.Stmt

	for !p.at(lexer.RBRACE) && !p.at(lexer.EOF) {
		switch p.peek().Kind {
		case lexer.LET:
			stmts = append(stmts, p.parseLetBinding())
		case lexer.REQUIRE:
			stmts = append(stmts, p.parseRequireStmt())
		default:
			pos := p.peek()
			expr := p.parseExpr(precOr)
			stmts = append(stmts, &ast.AssertExpr{Expr: expr, Pos: astPos(pos)})
		}
	}

	return stmts
}

// parseLetBinding parses: let name = expr or let (n1, n2) = expr
func (p *parser) parseLetBinding() *ast.LetBinding {
	start := p.advance() // consume LET

	var names []string

	if p.at(lexer.LPAREN) {
		p.advance() // consume LPAREN
		for !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
			nameTok, ok := p.expect(lexer.IDENT)
			if !ok {
				break
			}
			names = append(names, nameTok.Literal)
			if p.at(lexer.COMMA) {
				p.advance() // consume COMMA
			} else if !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
				p.addError(p.peek(), "expected comma between names")
			}
		}
		p.expect(lexer.RPAREN)
	} else {
		nameTok, ok := p.expect(lexer.IDENT)
		if !ok {
			return &ast.LetBinding{Pos: astPos(start)}
		}
		names = []string{nameTok.Literal}
	}

	if _, ok := p.expect(lexer.ASSIGN); !ok {
		return &ast.LetBinding{Names: names, Pos: astPos(start)}
	}

	expr := p.parseExpr(precOr)

	return &ast.LetBinding{Names: names, Expr: expr, Pos: astPos(start)}
}

// parseRequireStmt parses: require expr
func (p *parser) parseRequireStmt() *ast.RequireStmt {
	start := p.advance() // consume REQUIRE

	expr := p.parseExpr(precOr)

	return &ast.RequireStmt{Expr: expr, Pos: astPos(start)}
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
		if p.at(lexer.IS) && precPostfix >= minPrec {
			pos := p.advance() // consume IS
			targetTok := p.peek()
			var target ast.IsTarget
			switch targetTok.Kind {
			case lexer.OK:
				target = ast.IsOk
			case lexer.ERROR:
				target = ast.IsError
			default:
				p.addError(targetTok, "expected ok or error after 'is', got '%s'", targetTok.Kind)
				return left
			}
			p.advance() // consume ok/error
			left = &ast.IsExpr{Expr: left, Target: target, Pos: astPos(pos)}
			continue
		}

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
		if p.at(lexer.LPAREN) {
			return p.parseCallArgs(tok)
		}
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

// parseCallArgs parses the argument list of a function call.
// The function name token has already been consumed.
func (p *parser) parseCallArgs(nameTok lexer.Token) *ast.CallExpr {
	p.advance() // consume LPAREN

	var args []ast.Expr
	for !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
		before := p.pos
		arg := p.parseExpr(precOr)
		if p.pos == before {
			p.skipToParam()
			if p.at(lexer.COMMA) {
				p.advance() // consume COMMA
			}
			continue
		}
		args = append(args, arg)
		if p.at(lexer.COMMA) {
			p.advance() // consume COMMA
		} else if !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
			p.addError(p.peek(), "expected comma between arguments")
		}
	}

	p.expect(lexer.RPAREN)

	return &ast.CallExpr{Func: nameTok.Literal, Args: args, Pos: astPos(nameTok)}
}

// PRIVATE HELPERS

// skipToDecl advances past tokens until a declaration start, closing brace,
// or EOF is found. Used for error recovery.
func (p *parser) skipToDecl() {
	for {
		switch p.peek().Kind {
		case lexer.TYPE, lexer.FUNC, lexer.PREDICATE, lexer.PROPERTY, lexer.RBRACE, lexer.EOF:
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

func (p *parser) peekAt(offset int) lexer.Token {
	idx := p.pos + offset
	if idx >= len(p.tokens) {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[idx]
}

func (p *parser) addError(tok lexer.Token, format string, args ...any) {
	p.errors = append(p.errors, Error{
		Message: fmt.Sprintf(format, args...),
		Pos:     astPos(tok),
	})
}
