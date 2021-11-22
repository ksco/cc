package cc

type Scope struct {
	vars []*Object
	tags []*Type
}

type Parser struct {
	tokens    []*Token
	literals  []*Object
	scopes    []*Scope
	stackSize int
	pos       int
	strId     int
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{
		tokens: tokens,
		scopes: []*Scope{{}},
	}
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

func (p *Parser) EnterScope() {
	p.scopes = append([]*Scope{{}}, p.scopes...)
}

func (p *Parser) LeaveScope() {
	p.scopes = p.scopes[1:]
}

func (p *Parser) ScopeVars() []*Object {
	return p.scopes[0].vars
}

func (p *Parser) PushVarScope(o *Object) {
	p.scopes[0].vars = append(p.scopes[0].vars, o)
}

func (p *Parser) PushTagScope(t *Type) {
	p.scopes[0].tags = append([]*Type{t}, p.scopes[0].tags...)
}

func (p *Parser) AddLocals(locals ...*Object) {
	for _, l := range locals {
		l.Kind = OKLocal
		l.Local = &Local{}
		p.PushVarScope(l)
	}
}

func (p *Parser) AddGlobals(globals ...*Object) {
	for _, g := range globals {
		g.Kind = OKGlobal
		g.Global = &Global{}
		p.PushVarScope(g)
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

func (p *Parser) FindVariable(name string) *Object {
	for _, s := range p.scopes {
		for _, v := range s.vars {
			if name == v.Name {
				return v
			}
		}
	}
	return nil
}

func (p *Parser) FindTags(name string) *Type {
	for _, s := range p.scopes {
		for _, t := range s.tags {
			if t.Val.(*StructVal).Name != nil && name == t.Val.(*StructVal).Name.Val.(string) {
				return t
			}
		}
	}
	return nil
}

func (p *Parser) IsFunction() bool {
	pos := p.pos
	o, _ := p.Declarator(p.DeclSpec())
	p.MoveTo(pos)
	return o.Type.Kind == TYFunc
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
	f := &Function{}

	if p.Current().Equal(TKPunctuator, ";") {
		p.Consume(TKPunctuator, ";")
		f.IsDefinition = true
	} else {
		p.Consume(TKPunctuator, "{")
		p.AddLocals(params...)

		f.Body = p.Stmts()
		f.Locals = p.ScopeVars()
	}

	f.Params = params
	p.LeaveScope()
	return (&Object{
		Name:     o.Name,
		Kind:     OKFunction,
		Function: f,
	}).AlignLocals()
}

func (p *Parser) DeclSpec() *Type {
	if p.Current().Equal(TKKeyword, "long") {
		p.Consume(TKKeyword, "long")
		return LongType
	}
	if p.Current().Equal(TKKeyword, "int") {
		p.Consume(TKKeyword, "int")
		return IntType
	}
	if p.Current().Equal(TKKeyword, "short") {
		p.Consume(TKKeyword, "short")
		return ShortType
	}
	if p.Current().Equal(TKKeyword, "char") {
		p.Consume(TKKeyword, "char")
		return CharType
	}

	if p.Current().Equal(TKKeyword, "struct") {
		p.Consume(TKKeyword, "struct")
		return p.StructUnionDecl("struct")
	} else if p.Current().Equal(TKKeyword, "union") {
		p.Consume(TKKeyword, "union")
		return p.StructUnionDecl("union")
	}

	panic(p.Current().Errorf("type name expected"))
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
		return NewType(TYFunc, base, nil), p.FuncParams()
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
		return NewType(TYArray, t, tok.Val), nil
	}
	return base, nil
}

func (p *Parser) Declarator(base *Type) (*Object, []*Object) {
	for p.Current().Equal(TKPunctuator, "*") {
		p.Next()
		base = NewType(TYPtr, base, nil)
	}

	tok := p.Current()

	if tok.Equal(TKPunctuator, "(") {
		p.Next()
		nestedType, _ := p.Declarator(NewType(TYUnknown, nil, nil))
		p.Consume(TKPunctuator, ")")
		t, params := p.TypeSuffix(base)
		if nestedType.Type.Kind == TYUnknown {
			nestedType.Type = t
		} else {
			innerType := nestedType.Type
			for innerType.Base.Kind != TYUnknown {
				innerType = innerType.Base
			}
			innerType.Base = t
		}

		nestedType.Type.Resize()
		return nestedType, params
	}

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

		assigns = append(assigns, NewNode(NKExprStmt, &Unary{
			Expr: NewNode(NKAssign,
				&Binary{Lhs: NewNode(NKVariable, &Variable{Object: obj}, tok), Rhs: p.Assign()},
				tok)}, tok))
	}

	return NewNode(NKBlock, &Block{Stmts: assigns}, p.Current())
}

func (p *Parser) Stmt() *Node {
	cur := p.Current()
	if cur.Equal(TKKeyword, "return") {
		p.Next()

		expr := p.Expr()
		p.Consume(TKPunctuator, ";")
		return NewNode(NKReturn, &Unary{Expr: expr}, cur)
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
	return tok.Equal(TKKeyword, "long") ||
		tok.Equal(TKKeyword, "int") ||
		tok.Equal(TKKeyword, "short") ||
		tok.Equal(TKKeyword, "char") ||
		tok.Equal(TKKeyword, "struct") ||
		tok.Equal(TKKeyword, "union")
}

func (p *Parser) Stmts() *Node {
	var body []*Node
	tok := p.Current()
	for !p.Current().Equal(TKPunctuator, "}") {
		if p.IsTypeName() {
			body = append(body, p.Declaration())
		} else {
			body = append(body, p.Stmt())
		}
	}

	p.Next()

	return NewNode(NKBlock, &Block{Stmts: body}, tok)
}

func (p *Parser) ExprStmt() *Node {
	tok := p.Current()
	if p.Current().Equal(TKPunctuator, ";") {
		p.Next()
		return NewNode(NKBlock, &Block{}, tok)
	}

	tok = p.Current()
	expr := p.Expr()
	p.Consume(TKPunctuator, ";")
	return NewNode(NKExprStmt, &Unary{Expr: expr}, tok)
}

func (p *Parser) Expr() *Node {
	tok := p.Current()
	node := p.Assign()
	if p.Current().Equal(TKPunctuator, ",") {
		p.Next()
		return NewNode(NKComma, &Binary{Lhs: node, Rhs: p.Expr()}, tok)
	}

	return node
}

func (p *Parser) Assign() *Node {
	tok := p.Current()
	e := p.Equality()
	if p.Current().Equal(TKPunctuator, "=") {
		p.Next()
		e = NewNode(NKAssign, &Binary{Lhs: e, Rhs: p.Assign()}, tok)
	}

	return e
}

func (p *Parser) Equality() *Node {
	tok := p.Current()
	r := p.Relational()
	for true {
		if p.Current().Equal(TKPunctuator, "==") {
			p.Next()
			r = NewNode(NKEq, &Binary{Lhs: r, Rhs: p.Relational()}, tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, "!=") {
			p.Next()
			r = NewNode(NKNe, &Binary{Lhs: r, Rhs: p.Relational()}, tok)
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
			a = NewNode(NKLt, &Binary{Lhs: a, Rhs: p.Add()}, tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, "<=") {
			p.Next()
			a = NewNode(NKLe, &Binary{Lhs: a, Rhs: p.Add()}, tok)
			continue
		}

		if p.Current().Equal(TKPunctuator, ">") {
			p.Next()
			a = NewNode(NKLt, &Binary{Lhs: p.Add(), Rhs: a}, tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, ">=") {
			p.Next()
			a = NewNode(NKLe, &Binary{Lhs: p.Add(), Rhs: a}, tok)
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
			u = NewNode(NKMul, &Binary{Lhs: u, Rhs: p.Unary()}, tok)
			continue
		}
		if p.Current().Equal(TKPunctuator, "/") {
			p.Next()
			u = NewNode(NKDiv, &Binary{Lhs: u, Rhs: p.Unary()}, tok)
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
		return NewNode(NKNeg, &Unary{Expr: p.Unary()}, tok)
	}

	if tok.Equal(TKPunctuator, "*") {
		p.Next()
		return NewNode(NKDeRef, &Unary{Expr: p.Unary()}, tok)
	}

	if tok.Equal(TKPunctuator, "&") {
		p.Next()
		return NewNode(NKAddr, &Unary{Expr: p.Unary()}, tok)
	}

	return p.Postfix()
}

func (p *Parser) StructMembers() []*StructMember {
	ms := make([]*StructMember, 0)
	for !p.Current().Equal(TKPunctuator, "}") {
		base := p.DeclSpec()

		first := true
		for !p.Current().Equal(TKPunctuator, ";") {
			if !first {
				p.Consume(TKPunctuator, ",")
			}

			o, _ := p.Declarator(base)
			ms = append(ms, &StructMember{Type: o.Type, Name: o.Name})
			first = false
		}

		p.Next()
	}

	p.Next()
	return ms
}

func (p *Parser) StructUnionDecl(structOrUnion string) *Type {
	var tag *Token
	if p.Current().Kind == TKIdentifier {
		tag = p.Current()
		p.Next()
	}

	if tag != nil && !p.Current().Equal(TKPunctuator, "{") {
		t := p.FindTags(tag.Val.(string))
		if t == nil {
			panic(tag.Errorf("unknown %s type", structOrUnion))
		}
		return t
	}
	p.Consume(TKPunctuator, "{")

	ty := TYStruct
	if structOrUnion == "union" {
		ty = TYUnion
	}

	t := NewType(ty, nil, &StructVal{
		Members: p.StructMembers(),
		Name:    tag,
	})
	p.PushTagScope(t)
	return t
}

func (p *Parser) Postfix() *Node {
	n := p.Primary()

	for {
		if p.Current().Equal(TKPunctuator, "[") {
			p.Next()
			tok := p.Current()
			expr := p.Expr()
			p.Consume(TKPunctuator, "]")
			n = NewNode(NKDeRef, &Unary{Expr: NewNodeAdd(n, expr, tok)}, tok)
			continue
		}

		memberAccess := func() *Node {
			if n.Type.Kind != TYStruct && n.Type.Kind != TYUnion {
				panic(p.Current().Errorf("not a struct or union"))
			}

			for _, m := range n.Type.Val.(*StructVal).Members {
				if m.Name == p.Current().Lexeme {
					return NewNode(NKMember, &MemberAccess{Struct: n, Member: m}, p.Current())
				}
			}
			panic(p.Current().Errorf("no such member"))
		}
		if p.Current().Equal(TKPunctuator, ".") {
			p.Next()
			n = memberAccess()

			p.Next()
			continue
		}

		if p.Current().Equal(TKPunctuator, "->") {
			p.Next()
			n = NewNode(NKDeRef, &Unary{Expr: n}, p.Current())
			n = memberAccess()

			p.Next()
			continue
		}

		return n
	}
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
		return NewNode(NKStmtsExpr, block.Block, tok)
	}

	if tok.Equal(TKPunctuator, "(") {
		expr := p.Expr()
		p.Consume(TKPunctuator, ")")
		return expr
	}

	if tok.Equal(TKKeyword, "sizeof") {
		n := p.Unary()
		return NewNode(NKNum, &Number{Val: n.Type.Size}, tok)
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
		return NewNode(NKVariable, &Variable{Object: variable}, tok)
	}

	if tok.Kind == TKNumber {
		return NewNode(NKNum, &Number{Val: tok.Val.(int)}, tok)
	}

	if tok.Kind == TKString {
		o := &Object{
			Kind:   OKStringLiteral,
			Type:   tok.Val.(*String).Type,
			Global: &Global{Val: tok.Val.(*String).Val},
		}
		p.literals = append(p.literals, o)
		return NewNode(NKStringLiteral, &Variable{Object: o}, tok)
	}

	panic(tok.Errorf("expected an expression, got '%s' instead", tok.Lexeme))
}

func alignTo(n int, align int) int {
	return (n + align - 1) / align * align
}
