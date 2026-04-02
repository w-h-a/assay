package ast

// PropertyShape classifies a property for backend compatibility.
//
// Contractual properties assert a single predicate over function outputs
// (e.g., add(a, b) == add(b, a)). All backends can express these.
//
// Sequential properties chain function calls via let-bindings, guards,
// and terminal assertion (e.g., write then read-back). Property testing
// handles these naturally; formal provers require composed lemmas.
type PropertyShape string

const (
	Contractual PropertyShape = "contractual"
	Sequential  PropertyShape = "sequential"
)

// Expr is implemented by all expression nodes in the spec language.
//
//   - BinaryExpr:      a == b, x + y, p and q
//   - UnaryExpr:       not x
//   - CallExpr:        append(log, value)
//   - IdentExpr:       log, offset
//   - LiteralExpr:     42, "hello", true
//   - TupleExpr:       (a, b)
//   - IsExpr:          err is ok, err is error
//   - FieldAccessExpr: obj.field
type Expr interface {
	exprNode()
}

// BinaryExpr represents a binary operation (e.g., a == b, x + y).
type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
	Pos   Position
}

func (*BinaryExpr) exprNode() {}

// UnaryExpr represents a unary operation (e.g., not x).
type UnaryExpr struct {
	Op      string
	Operand Expr
	Pos     Position
}

func (*UnaryExpr) exprNode() {}

// CallExpr represents a function call (e.g., append(log, value)).
type CallExpr struct {
	Func string
	Args []Expr
	Pos  Position
}

func (*CallExpr) exprNode() {}

// IdentExpr represents a name reference.
type IdentExpr struct {
	Name string
	Pos  Position
}

func (*IdentExpr) exprNode() {}

// LiteralKind classifies a literal value.
type LiteralKind string

const (
	LiteralInt    LiteralKind = "int"
	LiteralFloat  LiteralKind = "float"
	LiteralString LiteralKind = "string"
	LiteralBool   LiteralKind = "bool"
	LiteralBytes  LiteralKind = "bytes"
)

// LiteralExpr represents a literal value.
type LiteralExpr struct {
	Value string
	Kind  LiteralKind
	Pos   Position
}

func (*LiteralExpr) exprNode() {}

// TupleExpr represents a tuple construction (e.g., (a, b)).
type TupleExpr struct {
	Elements []Expr
	Pos      Position
}

func (*TupleExpr) exprNode() {}

// IsTarget classifies the target of an is-expression.
type IsTarget string

const (
	IsOk    IsTarget = "ok"
	IsError IsTarget = "error"
)

// IsExpr represents an error check (e.g., err is ok, err is error).
type IsExpr struct {
	Expr   Expr
	Target IsTarget
	Pos    Position
}

func (*IsExpr) exprNode() {}

// FieldAccessExpr represents a field access (e.g., obj.field).
type FieldAccessExpr struct {
	Object Expr
	Field  string
	Pos    Position
}

func (*FieldAccessExpr) exprNode() {}

// Stmt is implemented by all statement nodes in a property body.
// A property body is a sequence of statements ending in an assertion:
//
//   - LetBinding:  let (offset, err) = append(log, value)
//   - RequireStmt: require err is ok  (guard — skips test case if false)
//   - AssertExpr:  result == value    (terminal assertion)
type Stmt interface {
	stmtNode()
}

// LetBinding binds a name or destructured tuple to an expression.
type LetBinding struct {
	Names []string
	Expr  Expr
	Pos   Position
}

func (*LetBinding) stmtNode() {}

// RequireStmt is a guard — skips the test case if false.
type RequireStmt struct {
	Expr Expr
	Pos  Position
}

func (*RequireStmt) stmtNode() {}

// AssertExpr is a bare expression used as the terminal assertion.
type AssertExpr struct {
	Expr Expr
	Pos  Position
}

func (*AssertExpr) stmtNode() {}

// GeneratorConstraint narrows the values produced for a quantified variable.
// A nil constraint on QuantifiedVar means the default generator for the type.
//
//   - RangeGen:   n: int in 1..100
//   - BuiltinGen: s: string in strings(1, 50)
//   - OneOfGen:   x: int in {1, 2, 3}
type GeneratorConstraint interface {
	genNode()
}

// RangeGen constrains a variable to a range (e.g., n: int in 1..100).
type RangeGen struct {
	Lo  Expr
	Hi  Expr
	Pos Position
}

func (*RangeGen) genNode() {}

// BuiltinGen constrains a variable using a builtin generator (e.g., strings(1, 50)).
type BuiltinGen struct {
	Name string
	Args []Expr
	Pos  Position
}

func (*BuiltinGen) genNode() {}

// OneOfGen constrains a variable to one of a set of values.
type OneOfGen struct {
	Values []Expr
	Pos    Position
}

func (*OneOfGen) genNode() {}

// QuantifiedVar is a variable bound by a forall clause.
type QuantifiedVar struct {
	Name      string
	Type      TypeExpr
	Generator GeneratorConstraint // nil means default generator
	Pos       Position
}

// ForallClause declares universally quantified variables.
type ForallClause struct {
	Vars []QuantifiedVar
	Pos  Position
}

// WhereClause is a precondition filter on generated inputs.
type WhereClause struct {
	Condition Expr
	Pos       Position
}

// PredicateDecl declares a named boolean predicate.
type PredicateDecl struct {
	Name   string
	Params []Param
	Body   Expr
	Pos    Position
}

func (*PredicateDecl) decNode() {}

// PropertyDecl declares a behavioral property.
type PropertyDecl struct {
	Name   string
	Forall ForallClause
	Where  *WhereClause // nil if no where clause
	Body   []Stmt
	Shape  PropertyShape
	Pos    Position
}

func (*PropertyDecl) decNode() {}
