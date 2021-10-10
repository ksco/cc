package cc

import "fmt"

type VarScope struct {
	next *VarScope
	o    *Object
}

type Scope struct {
	next *Scope
	vars *VarScope
}

type Parser struct {
	tokens    []*Token
	literals  []*Object
	scope     *Scope
	stackSize int
	pos       int
	strId     int
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{
		tokens: tokens,
		scope:  &Scope{},
	}
}

func (p *Parser) EnterScope() {
	p.scope = &Scope{next: p.scope}
}

func (p *Parser) LeaveScope() {
	p.scope = p.scope.next
}

func (p *Parser) ScopeVars() []*Object {
	vars := make([]*Object, 0)
	for v := p.scope.vars; v != nil; v = v.next {
		vars = append(vars, v.o)
	}
	return vars
}

func (p *Parser) PushScope(o *Object) {
	p.scope.vars = &VarScope{next: p.scope.vars, o: o}
}

func (p *Parser) AddLocals(locals ...*Object) {
	for _, l := range locals {
		l.Val = &Local{}
		p.PushScope(l)
	}
}

func (p *Parser) AddGlobals(globals ...*Object) {
	for _, g := range globals {
		g.Val = &Global{}
		p.PushScope(g)
	}
}

func (p *Parser) ReachedEOF() bool {
	if p.Current().Equal(TKEof, "") {
		return true
	}

	return false
}

func (p *Parser) Current() *Token {
	return p.tokens[p.pos]
}

func (p *Parser) Next() {
	p.pos += 1
}

func (p *Parser) MoveTo(pos int) {
	p.pos = pos
}

func (p *Parser) Consume(kind TokenKind, lexeme string) {
	cur := p.Current()
	if !cur.Equal(kind, lexeme) {
		panic(cur.Errorf("expected '%s', got '%s' instead", lexeme, cur.Lexeme))
	}

	p.Next()
}

func (p *Parser) FindVariable(name string) (o *Object) {
	for s := p.scope; s != nil; s = s.next {
		for vs := s.vars; vs != nil; vs = vs.next {
			if name == vs.o.Name {
				return vs.o
			}
		}
	}
	return nil
}

func (p *Parser) Parse() (objects []*Object, err error) {
	defer func() {
		var r interface{}
		if r = recover(); r == nil {
			return
		}

		var ok bool
		if err, ok = r.(error); !ok {
			panic(r)
		}
	}()

	for !p.ReachedEOF() {
		if p.IsFunction() {
			objects = append(objects, p.FuncDef())
			continue
		}
		objects = append(objects, p.GlobalVariables()...)
	}

	objects = append(objects, p.literals...)

	return
}

func (p *Parser) IsFunction() bool {
	pos := p.pos
	o, _ := p.Declarator(p.DeclSpec())
	p.MoveTo(pos)
	return o.Type.Kind == TypeKindFunc
}

func (p *Parser) GlobalVariables() []*Object {
	base := p.DeclSpec()
	first := true
	for !p.Current().Equal(TKPunctuator, ";") {
		if !first {
			p.Consume(TKPunctuator, ",")
		}
		first = false
		o, _ := p.Declarator(base)
		p.AddGlobals(o)
	}
	p.Next()

	return p.ScopeVars()
}

func (p *Parser) FuncDef() *Object {
	p.EnterScope()
	o, params := p.Declarator(p.DeclSpec())

	p.Consume(TKPunctuator, "{")

	p.AddLocals(params...)
	body := p.Stmts()
	body.AddType()
	f := &Function{
		Body:   body,
		Locals: p.ScopeVars(),
		Params: params,
	}
	p.LeaveScope()
	return (&Object{
		Name: o.Name,
		Val:  f,
	}).AlignLocals()
}

func (p *Parser) DeclSpec() *Type {
	if p.Current().Equal(TKKeyword, "char") {
		p.Consume(TKKeyword, "char")
		return CharType
	}
	p.Consume(TKKeyword, "int")
	return IntType
}

func (p *Parser) FuncParams() []*Object {
	params := make([]*Object, 0)
	first := true
	for !p.Current().Equal(TKPunctuator, ")") {
		if !first {
			p.Consume(TKPunctuator, ",")
		}
		first = false
		o, _ := p.Declarator(p.DeclSpec())
		params = append(params, o)
	}
	p.Next()
	return params
}

func (p *Parser) TypeSuffix(base *Type) (*Type, []*Object) {
	if p.Current().Equal(TKPunctuator, "(") {
		p.Next()
		return NewType(TypeKindFunc, base, nil), p.FuncParams()
	}
	if p.Current().Equal(TKPunctuator, "[") {
		p.Next()

		tok := p.Current()
		if tok.Kind != TKNumber {
			panic(tok.Errorf("expected a number, got '%s' instead", tok.Lexeme))
		}
		p.Next()
		p.Consume(TKPunctuator, "]")
		t, _ := p.TypeSuffix(base)
		return NewType(TypeKindArray, t, tok.Val), nil
	}
	return base, nil
}

func (p *Parser) Declarator(base *Type) (*Object, []*Object) {
	for p.Current().Equal(TKPunctuator, "*") {
		p.Next()
		base = NewType(TypeKindPtr, base, nil)
	}

	tok := p.Current()
	if tok.Kind != TKIdentifier {
		panic(tok.Errorf("expected a variable name, got '%s' instead", tok.Lexeme))
	}
	p.Next()

	t, params := p.TypeSuffix(base)
	return &Object{Name: tok.Val.(string), Type: t}, params
}

func (p *Parser) Declaration() *Node {
	base := p.DeclSpec()

	first := true
	assigns := make([]*Node, 0)
	for !p.Current().Equal(TKPunctuator, ";") {
		if !first {
			p.Consume(TKPunctuator, ",")
		}
		first = false

		obj, _ := p.Declarator(base)
		p.AddLocals(obj)

		tok := p.Current()
		if !tok.Equal(TKPunctuator, "=") {
			continue
		}
		p.Next()

		assigns = append(assigns, NewNode(
			NKExprStmt,
			NewNode(
				NKAssign,
				&BinaryExpr{
					Lhs: NewNode(NKVariable, obj, tok),
					Rhs: p.Expr(),
				}, tok),
			tok),
		)
	}

	return NewNode(NKBlock, assigns, p.Current())
}

func (p *Parser) Stmt() *Node {
	cur := p.Current()
	if cur.Equal(TKKeyword, "return") {
		p.Next()

		expr := p.Expr()
		p.Consume(TKPunctuator, ";")
		return NewNode(NKReturn, expr, cur)
	}

	if cur.Equal(TKKeyword, "if") {
		p.Next()
		ifClause := &IfClause{}

		p.Consume(TKPunctuator, "(")
		ifClause.Cond = p.Expr()
		p.Consume(TKPunctuator, ")")
		ifClause.Then = p.Stmt()

		cur = p.Current()
		if cur.Equal(TKKeyword, "else") {
			p.Next()
			ifClause.Else = p.Stmt()
		}

		return NewNode(NKIf, ifClause, cur)
	}

	if cur.Equal(TKKeyword, "for") {
		p.Next()
		forClause := &ForClause{}
		p.Consume(TKPunctuator, "(")
		forClause.Init = p.ExprStmt()
		if !p.Current().Equal(TKPunctuator, ";") {
			forClause.Cond = p.Expr()
		}
		p.Consume(TKPunctuator, ";")
		if !p.Current().Equal(TKPunctuator, ")") {
			forClause.Increment = p.Expr()
		}
		p.Consume(TKPunctuator, ")")
		forClause.Body = p.Stmt()
		return NewNode(NKFor, forClause, cur)
	}

	if cur.Equal(TKKeyword, "while") {
		p.Next()
		forClause := &ForClause{}
		p.Consume(TKPunctuator, "(")
		forClause.Cond = p.Expr()
		p.Consume(TKPunctuator, ")")
		forClause.Body = p.Stmt()
		return NewNode(NKFor, forClause, cur)
	}

	if cur.Equal(TKPunctuator, "{") {
		p.Next()
		p.EnterScope()
		stmts := p.Stmts()
		p.LeaveScope()
		return stmts
	}

	return p.ExprStmt()
}

func (p *Parser) IsTypeName() bool {
	tok := p.Current()
	return tok.Equal(TKKeyword, "int") || tok.Equal(TKKeyword, "char")
}

func (p *Parser) Stmts() *Node {
	var body []*Node
	for !p.Current().Equal(TKPunctuator, "}") {
		if p.IsTypeName() {
			body = append(body, p.Declaration())
		} else {
			body = append(body, p.Stmt())
		}
	}

	p.Next()

	return &Node{
		Kind: NKBlock,
		Val:  body,
	}
}

func (p *Parser) ExprStmt() *Node {
	tok := p.Current()
	if p.Current().Equal(TKPunctuator, ";") {
		p.Next()
		return NewNode(NKBlock, make([]*Node, 0), tok)

	}

	tok = p.Current()
	expr := p.Expr()
	p.Consume(TKPunctuator, ";")
	return NewNode(NKExprStmt, expr, tok)
}

func (p *Parser) Expr() *Node {
	return p.Assign()
}

func (p *Parser) Assign() *Node {
	tok := p.Current()
	e := p.Equality()
	if p.Current().Equal(TKPunctuator, "=") {
		p.Next()
		e = NewNode(NKAssign, &BinaryExpr{Lhs: e, Rhs: p.Assign()}, tok)
	}

	return e
}

func (p *Parser) Equality() *Node {
	tok := p.Current()
	r := p.Relational()
	for true {
		if p.Current().Equal(TKPunctuator, "==") {
			p.Next()
			r = NewNode(NKEq, &BinaryExpr{Lhs: r, Rhs: p.Relational()}, tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, "!=") {
			p.Next()
			r = NewNode(NKNe, &BinaryExpr{Lhs: r, Rhs: p.Relational()}, tok)
			continue
		}

		return r
	}

	// Unreachable
	return nil
}

func (p *Parser) Relational() *Node {
	tok := p.Current()
	a := p.Add()
	for true {
		if p.Current().Equal(TKPunctuator, "<") {
			p.Next()
			a = NewNode(NKLt, &BinaryExpr{Lhs: a, Rhs: p.Add()}, tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, "<=") {
			p.Next()
			a = NewNode(NKLe, &BinaryExpr{Lhs: a, Rhs: p.Add()}, tok)
			continue
		}

		if p.Current().Equal(TKPunctuator, ">") {
			p.Next()
			a = NewNode(NKLt, &BinaryExpr{Lhs: p.Add(), Rhs: a}, tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, ">=") {
			p.Next()
			a = NewNode(NKLe, &BinaryExpr{Lhs: p.Add(), Rhs: a}, tok)
			continue
		}

		return a
	}

	// Unreachable
	return nil
}

func (p *Parser) Add() *Node {
	tok := p.Current()
	m := p.Mul()
	for true {
		if p.Current().Equal(TKPunctuator, "+") {
			p.Next()
			m = NewNodeAdd(m, p.Mul(), tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, "-") {
			p.Next()
			m = NewNodeSub(m, p.Mul(), tok)
			continue
		}

		return m
	}

	// Unreachable
	return nil
}

func (p *Parser) Mul() *Node {
	tok := p.Current()
	u := p.Unary()
	for true {
		if p.Current().Equal(TKPunctuator, "*") {
			p.Next()
			u = NewNode(NKMul, &BinaryExpr{Lhs: u, Rhs: p.Unary()}, tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, "/") {
			p.Next()
			u = NewNode(NKDiv, &BinaryExpr{Lhs: u, Rhs: p.Unary()}, tok)
			continue
		}

		return u
	}

	// Unreachable
	return nil
}

func (p *Parser) Unary() *Node {
	tok := p.Current()
	if tok.Equal(TKPunctuator, "+") {
		p.Next()
		return p.Unary()
	}

	if tok.Equal(TKPunctuator, "-") {
		p.Next()
		return NewNode(NKNeg, p.Unary(), tok)
	}

	if tok.Equal(TKPunctuator, "*") {
		p.Next()
		return NewNode(NKDeRef, p.Unary(), tok)
	}

	if tok.Equal(TKPunctuator, "&") {
		p.Next()
		return NewNode(NKAddr, p.Unary(), tok)
	}

	return p.Postfix()
}

func (p *Parser) Postfix() *Node {
	n := p.Primary()
	n.AddType()
	for p.Current().Equal(TKPunctuator, "[") {
		p.Next()
		tok := p.Current()
		expr := p.Expr()
		p.Consume(TKPunctuator, "]")
		n = NewNode(NKDeRef, NewNodeAdd(n, expr, tok), tok)
	}

	return n
}

func (p *Parser) FuncCall(tok *Token) *Node {
	args := make([]*Node, 0)
	first := true
	for !p.Current().Equal(TKPunctuator, ")") {
		if !first {
			p.Consume(TKPunctuator, ",")
		}
		first = false
		args = append(args, p.Assign())
	}
	p.Next()
	return NewNode(NKFuncCall, &FuncCall{
		Name: tok.Val.(string),
		Args: args,
	}, tok)
}

func (p *Parser) Primary() *Node {
	tok := p.Current()
	p.Next()

	if tok.Equal(TKPunctuator, "(") && p.Current().Equal(TKPunctuator, "{") {
		p.Next()
		p.EnterScope()
		block := p.Stmts()
		p.LeaveScope()
		p.Consume(TKPunctuator, ")")
		return NewNode(NKStmtExpr, block, tok)
	}

	if tok.Equal(TKPunctuator, "(") {
		expr := p.Expr()
		p.Consume(TKPunctuator, ")")
		return expr
	}

	if tok.Equal(TKKeyword, "sizeof") {
		n := p.Unary()
		n.AddType()
		return NewNode(NKNum, n.Type.Size, tok)
	}

	if tok.Kind == TKIdentifier {
		if p.Current().Equal(TKPunctuator, "(") {
			p.Next()
			return p.FuncCall(tok)
		}
		variable := p.FindVariable(tok.Val.(string))
		if variable == nil {
			panic(tok.Errorf("undefined variable '%s'", tok.Val.(string)))
		}
		return NewNode(NKVariable, variable, tok)
	}

	if tok.Kind == TKNumber {
		return NewNode(NKNum, tok.Val, tok)
	}

	if tok.Kind == TKString {
		o := &Object{
			Name: p.newStrId(),
			Type: tok.Val.(*String).Type,
			Val:  &Global{Val: tok.Val.(*String).Val},
		}
		p.literals = append(p.literals, o)
		return NewNode(NKVariable, o, tok)
	}

	panic(tok.Errorf("expected an expression, got '%s' instead", tok.Lexeme))
}

func (p *Parser) newStrId() (s string) {
	s = fmt.Sprintf(".L.str.%d", p.strId)
	p.strId++
	return
}

func alignTo(n int, align int) int {
	return (n + align - 1) / align * align
}