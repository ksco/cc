package cc

type NodeKind int

const (
	NKAdd       NodeKind = iota // +
	NKSub                       // -
	NKMul                       // *
	NKDiv                       // /
	NKNeg                       // unary -
	NKEq                        // ==
	NKNe                        // !=
	NKLt                        // <
	NKLe                        // <=
	NKAssign                    // =
	NKComma                     // ,
	NKMember                    // .
	NKAddr                      // unary &
	NKDeRef                     // unary *
	NKReturn                    // "return"
	NKIf                        // "if"
	NKFor                       // "for", "while"
	NKBlock                     // { ... }
	NKFuncCall                  // function call
	NKExprStmt                  // expression stmt
	NKStmtsExpr                 // stmts expression
	NKVariable                  // Variable
	NKNum                       // integer
)

type StructMember struct {
	Type   *Type
	Name   string
	Offset int
}

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

type Binary struct {
	Lhs *Node
	Rhs *Node
}

type Unary struct {
	Expr *Node
}

type FuncCall struct {
	Name string
	Args []*Node
}

type MemberAccess struct {
	Struct *Node
	Member *StructMember
}

type Block struct {
	Stmts []*Node
}

type Variable struct {
	Object *Object
}

type Number struct {
	Val int
}

type Node struct {
	Kind NodeKind
	Type *Type
	Tok  *Token

	// Only one of the following fields should be set.
	Num          *Number
	Variable     *Variable
	IfClause     *IfClause
	ForClause    *ForClause
	Binary       *Binary
	Unary        *Unary
	FuncCall     *FuncCall
	MemberAccess *MemberAccess
	Block        *Block
}

type NodeVal interface {
	IsNodeVal()
}

func (*Number) IsNodeVal()       {}
func (*Variable) IsNodeVal()     {}
func (*IfClause) IsNodeVal()     {}
func (*ForClause) IsNodeVal()    {}
func (*Binary) IsNodeVal()       {}
func (*Unary) IsNodeVal()        {}
func (*FuncCall) IsNodeVal()     {}
func (*MemberAccess) IsNodeVal() {}
func (*Block) IsNodeVal()        {}

func NewNode(kind NodeKind, val NodeVal, tok *Token) *Node {
	n := &Node{Kind: kind, Tok: tok}
	switch kind {
	case NKAdd, NKSub, NKMul, NKDiv, NKEq, NKNe, NKLt, NKLe, NKAssign, NKComma:
		n.Binary = val.(*Binary)
	case NKNeg, NKAddr, NKDeRef, NKReturn, NKExprStmt:
		n.Unary = val.(*Unary)
	case NKMember:
		n.MemberAccess = val.(*MemberAccess)
	case NKIf:
		n.IfClause = val.(*IfClause)
	case NKFor:
		n.ForClause = val.(*ForClause)
	case NKBlock, NKStmtsExpr:
		n.Block = val.(*Block)
	case NKFuncCall:
		n.FuncCall = val.(*FuncCall)
	case NKNum:
		n.Num = val.(*Number)
	case NKVariable:
		n.Variable = val.(*Variable)
	}
	n.addType()
	return n
}

func NewNodeAdd(lhs *Node, rhs *Node, tok *Token) *Node {
	if lhs.Type.IsInteger() && rhs.Type.IsInteger() {
		return NewNode(NKAdd, &Binary{Lhs: lhs, Rhs: rhs}, tok)
	}

	if lhs.Type.Base != nil && rhs.Type.Base != nil {
		panic(tok.Errorf("invalid operands"))
	}

	if rhs.Type.Base != nil {
		rhs, lhs = lhs, rhs
	}

	rhs = NewNode(NKMul, &Binary{
		Lhs: rhs,
		Rhs: NewNode(NKNum, &Number{Val: lhs.Type.Base.Size}, tok),
	}, tok)

	n := NewNode(NKAdd, &Binary{Lhs: lhs, Rhs: rhs}, tok)
	n.addType()
	return n
}

func NewNodeSub(lhs *Node, rhs *Node, tok *Token) *Node {
	lhs.addType()
	rhs.addType()
	if lhs.Type.IsInteger() && rhs.Type.IsInteger() {
		return NewNode(NKSub, &Binary{Lhs: lhs, Rhs: rhs}, tok)
	}

	if lhs.Type.Base != nil && rhs.Type.IsInteger() {
		rhs = NewNode(NKMul, &Binary{
			Lhs: rhs,
			Rhs: NewNode(NKNum, &Number{Val: lhs.Type.Base.Size}, tok),
		}, tok)

		return NewNode(NKSub, &Binary{Lhs: lhs, Rhs: rhs}, tok)
	}

	if lhs.Type.Base != nil && rhs.Type.Base != nil {
		return NewNode(
			NKDiv,
			&Binary{
				Lhs: NewNode(NKSub, &Binary{Lhs: lhs, Rhs: rhs}, tok),
				Rhs: NewNode(NKNum, &Number{Val: lhs.Type.Base.Size}, tok),
			},
			tok,
		)
	}

	panic(tok.Errorf("invalid operands"))
}

func (n *Node) addType() {
	if n.Type != nil {
		return
	}

	switch n.Kind {
	case NKNeg, NKAddr, NKDeRef, NKReturn, NKExprStmt:
		if node := n.Unary.Expr; node != nil {
			node.addType()
		}
	case NKAdd, NKSub, NKMul, NKDiv,
		NKEq, NKNe, NKLt, NKLe, NKAssign:
		n.Binary.Lhs.addType()
		n.Binary.Rhs.addType()
	case NKBlock:
		for _, node := range n.Block.Stmts {
			node.addType()
		}
	case NKIf:
		if n.IfClause.Cond != nil {
			n.IfClause.Cond.addType()
		}
		if n.IfClause.Then != nil {
			n.IfClause.Then.addType()
		}
		if n.IfClause.Else != nil {
			n.IfClause.Else.addType()
		}
	case NKFor:
		if n.ForClause.Init != nil {
			n.ForClause.Init.addType()
		}
		if n.ForClause.Cond != nil {
			n.ForClause.Cond.addType()
		}
		if n.ForClause.Increment != nil {
			n.ForClause.Increment.addType()
		}
		if n.ForClause.Body != nil {
			n.ForClause.Body.addType()
		}
	}

	switch n.Kind {
	case NKAdd, NKSub, NKMul, NKDiv, NKAssign:
		n.Type = n.Binary.Lhs.Type
		if n.Kind == NKSub &&
			n.Binary.Lhs.Type.Kind == TYPtr &&
			n.Binary.Rhs.Type.Kind == TYPtr {
			n.Type = IntType
		}
	case NKComma:
		n.Binary.Lhs.addType()
		n.Binary.Rhs.addType()
		n.Type = n.Binary.Rhs.Type
	case NKNeg:
		n.Type = n.Unary.Expr.Type
	case NKEq, NKNe, NKLt, NKLe, NKNum, NKFuncCall:
		n.Type = IntType
	case NKVariable:
		n.Type = n.Variable.Object.Type
	case NKMember:
		n.Type = n.MemberAccess.Member.Type
	case NKAddr:
		if n.Unary.Expr.Type.Kind == TYArray {
			n.Type = NewType(TYPtr, n.Unary.Expr.Type.Base, nil)
		} else {
			n.Type = NewType(TYPtr, n.Unary.Expr.Type, nil)
		}
	case NKDeRef:
		if n.Unary.Expr.Type.Base == nil {
			panic(n.Tok.Errorf("invalid pointer dereference"))
		}
		n.Type = n.Unary.Expr.Type.Base
	case NKStmtsExpr:
		if len(n.Block.Stmts) == 0 {
			panic(n.Tok.Errorf("statement expression returning void is not supported"))
		}
		last := n.Block.Stmts[len(n.Block.Stmts)-1]
		if last.Kind != NKExprStmt {
			panic(n.Tok.Errorf("statement expression returning void is not supported"))
		}
		n.Type = last.Unary.Expr.Type
	}
}
