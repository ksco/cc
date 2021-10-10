package cc

type NodeKind int

const (
	NKAdd      NodeKind = iota // +
	NKSub                      // -
	NKMul                      // *
	NKDiv                      // /
	NKNeg                      // unary -
	NKEq                       // ==
	NKNe                       // !=
	NKLt                       // <
	NKLe                       // <=
	NKAssign                   // =
	NKAddr                     // unary &
	NKDeRef                    // unary *
	NKReturn                   // "return"
	NKIf                       // "if"
	NKFor                      // "for", "while"
	NKBlock                    // { ... }
	NKFuncCall                 // function call
	NKExprStmt                 // expression stmt
	NKStmtExpr                 // stmt expression
	NKVariable                 // Variable
	NKNum                      // integer
)

type IfClause struct {
	Cond *Node
	Then *Node
	Else *Node
}

type ForClause struct {
	Init      *Node
	Cond      *Node
	Increment *Node
	Body      *Node
}

type BinaryExpr struct {
	Lhs *Node
	Rhs *Node
}

type FuncCall struct {
	Name string
	Args []*Node
}

type Node struct {
	Kind NodeKind
	Type *Type
	Tok  *Token
	Val  interface{}
}

func NewNode(kind NodeKind, val interface{}, tok *Token) *Node {
	return &Node{
		Kind: kind,
		Tok:  tok,
		Val:  val,
	}
}

func NewNodeAdd(lhs *Node, rhs *Node, tok *Token) *Node {
	lhs.AddType()
	rhs.AddType()

	if lhs.Type.IsInteger() && rhs.Type.IsInteger() {
		return NewNode(NKAdd, &BinaryExpr{Lhs: lhs, Rhs: rhs}, tok)
	}

	if lhs.Type.Base != nil && rhs.Type.Base != nil {
		panic(tok.Errorf("invalid operands"))
	}

	if rhs.Type.Base != nil {
		rhs, lhs = lhs, rhs
	}

	rhs = NewNode(NKMul, &BinaryExpr{
		Lhs: rhs,
		Rhs: NewNode(NKNum, lhs.Type.Base.Size, tok),
	}, tok)

	return NewNode(NKAdd, &BinaryExpr{Lhs: lhs, Rhs: rhs}, tok)
}

func NewNodeSub(lhs *Node, rhs *Node, tok *Token) *Node {
	lhs.AddType()
	rhs.AddType()
	if lhs.Type.IsInteger() && rhs.Type.IsInteger() {
		return NewNode(NKSub, &BinaryExpr{Lhs: lhs, Rhs: rhs}, tok)
	}

	if lhs.Type.Base != nil && rhs.Type.IsInteger() {
		rhs = NewNode(NKMul, &BinaryExpr{
			Lhs: rhs,
			Rhs: NewNode(NKNum, lhs.Type.Base.Size, tok),
		}, tok)

		return NewNode(NKSub, &BinaryExpr{Lhs: lhs, Rhs: rhs}, tok)
	}

	if lhs.Type.Base != nil && rhs.Type.Base != nil {
		return NewNode(
			NKDiv,
			&BinaryExpr{
				Lhs: NewNode(NKSub, &BinaryExpr{Lhs: lhs, Rhs: rhs}, tok),
				Rhs: NewNode(NKNum, lhs.Type.Base.Size, tok),
			},
			tok,
		)
	}

	panic(tok.Errorf("invalid operands"))
}

func (n *Node) AddType() {
	switch n.Kind {
	case NKNeg, NKAddr, NKDeRef, NKReturn, NKExprStmt:
		// Unary
		if node := n.Val.(*Node); node != nil {
			node.AddType()
		}
	case NKAdd, NKSub, NKMul, NKDiv,
		NKEq, NKNe, NKLt, NKLe, NKAssign:
		// Binary
		expr := n.Val.(*BinaryExpr)
		expr.Lhs.AddType()
		expr.Rhs.AddType()
	case NKBlock:
		nodes := n.Val.([]*Node)
		for _, node := range nodes {
			node.AddType()
		}
	case NKIf:
		ifClause := n.Val.(*IfClause)
		if ifClause.Cond != nil {
			ifClause.Cond.AddType()
		}
		if ifClause.Then != nil {
			ifClause.Then.AddType()
		}
		if ifClause.Else != nil {
			ifClause.Else.AddType()
		}
	case NKFor:
		forClause := n.Val.(*ForClause)
		if forClause.Init != nil {
			forClause.Init.AddType()
		}
		if forClause.Cond != nil {
			forClause.Cond.AddType()
		}
		if forClause.Increment != nil {
			forClause.Increment.AddType()
		}
		if forClause.Body != nil {
			forClause.Body.AddType()
		}
	}

	switch n.Kind {
	case NKAdd, NKSub, NKMul, NKDiv, NKAssign:
		node := n.Val.(*BinaryExpr)
		n.Type = node.Lhs.Type
		if n.Kind == NKSub && node.Lhs.Type.Kind == TypeKindPtr && node.Rhs.Type.Kind == TypeKindPtr {
			n.Type = IntType
		}
	case NKNeg:
		n.Type = n.Val.(*Node).Type
	case NKEq, NKNe, NKLt, NKLe, NKNum, NKFuncCall:
		n.Type = IntType
	case NKVariable:
		n.Type = n.Val.(*Object).Type
	case NKAddr:
		if n.Val.(*Node).Type.Kind == TypeKindArray {
			n.Type = NewType(TypeKindPtr, n.Val.(*Node).Type.Base, nil)
		} else {
			n.Type = NewType(TypeKindPtr, n.Val.(*Node).Type, nil)
		}
	case NKDeRef:
		if n.Val.(*Node).Type.Base == nil {
			panic(n.Tok.Errorf("invalid pointer dereference"))
		}
		n.Type = n.Val.(*Node).Type.Base
	case NKStmtExpr:
		n.Val.(*Node).AddType()
		stmts := n.Val.(*Node).Val.([]*Node)
		if len(stmts) == 0 {
			panic(n.Tok.Errorf("statement expression returning void is not supported"))
		}
		last := stmts[len(stmts)-1]
		if last.Kind != NKExprStmt {
			panic(n.Tok.Errorf("statement expression returning void is not supported"))
		}
		n.Type = last.Val.(*Node).Type
	}
}
