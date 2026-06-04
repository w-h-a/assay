package parser

import (
	"fmt"

	"github.com/w-h-a/assay/internal/binding/ast"
	"github.com/w-h-a/assay/internal/binding/lexer"
)

// Error represents a parsing error with source position.
type Error struct {
	Message string
	Pos     ast.Position
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// Parse delegates lexing and then parses a token stream into a binding AST.
func Parse(source, file string) (*ast.BindingDecl, []Error) {
	tokens := lexer.Lex(source, file)
	p := &parser{tokens: tokens}
	decl := p.parseBinding()
	return decl, p.errors
}

// parser is a recursive-descent parser that walks a pre-lexed token slice.
// Errors are collected rather than halting, so a single parse pass can
// report multiple problems.
type parser struct {
	tokens []lexer.Token
	pos    int
	errors []Error
}

// parseBinding parses the top-level bind block: bind "name" { <statements> }.
func (p *parser) parseBinding() *ast.BindingDecl {
	start := p.peek()

	if _, ok := p.expect(lexer.BIND); !ok {
		return &ast.BindingDecl{Pos: astPos(start)}
	}

	nameTok, ok := p.expect(lexer.STRING_LIT)
	if !ok {
		return &ast.BindingDecl{Pos: astPos(start)}
	}

	if _, ok := p.expect(lexer.LBRACE); !ok {
		return &ast.BindingDecl{Name: nameTok.Literal, Pos: astPos(start)}
	}

	decl := &ast.BindingDecl{
		Name: nameTok.Literal,
		Pos:  astPos(start),
	}

	for !p.at(lexer.RBRACE) && !p.at(lexer.EOF) {
		p.parseStatement(decl)
	}

	p.expect(lexer.RBRACE)
	if !p.at(lexer.EOF) {
		tok := p.peek()
		p.addError(tok, "unexpected token %s after bind declaration", tok.Kind)
	}

	return decl
}

// parseStatement dispatches on the current token to the appropriate
// statement parser inside a bind block.
func (p *parser) parseStatement(decl *ast.BindingDecl) {
	switch p.peek().Kind {
	case lexer.TARGET:
		p.parseTarget(decl)
	case lexer.PACKAGE:
		p.parsePackage(decl)
	case lexer.TYPE:
		decl.TypeMappings = append(decl.TypeMappings, p.parseTypeMapping())
	case lexer.FUNC:
		decl.FuncMappings = append(decl.FuncMappings, p.parseFuncMapping())
	default:
		tok := p.peek()
		p.addError(tok, "expected target, package, type, or func, got %s", tok.Kind)
		p.skipToStatement()
	}
}

// parseTarget parses 'target <ident>'. The identifier is recorded verbatim;
// the resolver decides which targets are supported.
func (p *parser) parseTarget(decl *ast.BindingDecl) {
	p.advance() // consume TARGET
	tok, ok := p.expect(lexer.IDENT)
	if !ok {
		return
	}
	decl.Target = tok.Literal
}

// parsePackage parses 'package "<path>"'. The path is recorded verbatim;
// the resolver loads and validates the package.
func (p *parser) parsePackage(decl *ast.BindingDecl) {
	p.advance() // consume PACKAGE
	tok, ok := p.expect(lexer.STRING_LIT)
	if !ok {
		return
	}
	decl.PackagePath = tok.Literal
}

// parseTypeMapping parses 'type SpecName = Qualifier.Name'.
func (p *parser) parseTypeMapping() ast.TypeMapping {
	start := p.advance() // consume TYPE
	tm := ast.TypeMapping{Pos: astPos(start)}

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return tm
	}
	tm.SpecName = nameTok.Literal

	if _, ok := p.expect(lexer.ASSIGN); !ok {
		return tm
	}

	qual, name, ok := p.parseDottedRef()
	if !ok {
		return tm
	}
	tm.Qualifier = qual
	tm.Name = name
	return tm
}

// parseFuncMapping parses 'func SpecName = Qualifier.Name'.
func (p *parser) parseFuncMapping() ast.FuncMapping {
	start := p.advance() // consume FUNC
	fm := ast.FuncMapping{Pos: astPos(start)}

	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return fm
	}
	fm.SpecName = nameTok.Literal

	if _, ok := p.expect(lexer.ASSIGN); !ok {
		return fm
	}

	qual, name, ok := p.parseDottedRef()
	if !ok {
		return fm
	}
	fm.Qualifier = qual
	fm.Name = name
	return fm
}

// parseDottedRef parses 'IDENT . IDENT' and returns (qualifier, name).
func (p *parser) parseDottedRef() (string, string, bool) {
	qualTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return "", "", false
	}
	if _, ok := p.expect(lexer.DOT); !ok {
		return "", "", false
	}
	nameTok, ok := p.expect(lexer.IDENT)
	if !ok {
		return "", "", false
	}
	return qualTok.Literal, nameTok.Literal, true
}

// skipToStatement advances past tokens until a statement-starting keyword
// or the closing brace at the bind level is found. Brace depth is tracked
// so nested blocks inside malformed input are consumed entirely.
func (p *parser) skipToStatement() {
	depth := 0
	for {
		switch p.peek().Kind {
		case lexer.LBRACE:
			depth++
			p.advance()
		case lexer.RBRACE:
			if depth > 0 {
				depth--
				p.advance()
				continue
			}
			return
		case lexer.TARGET, lexer.PACKAGE, lexer.TYPE, lexer.FUNC:
			if depth == 0 {
				return
			}
			p.advance()
		case lexer.EOF:
			return
		default:
			p.advance()
		}
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
