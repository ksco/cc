package cc

import (
	"errors"
	"fmt"
	"io"
)

type CodeGenerator struct {
	objects    []*Object
	depth      int
	labelCount int
	writer     io.Writer
}

func (g *CodeGenerator) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(g.writer, format, a...)
}

func NewCodeGenerator(w io.Writer, objects []*Object) *CodeGenerator {
	return &CodeGenerator{writer: w, objects: objects}
}

func (g *CodeGenerator) LabelCount() int {
	g.labelCount += 1
	return g.labelCount
}

func (g *CodeGenerator) Gen() (err error) {
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
	g.GenCode()
	g.GenData()
	return
}

func (g *CodeGenerator) GenData() {
	for _, o := range g.objects {
		global, ok := o.Val.(*Global)
		if !ok {
			continue
		}

		if global.Val != nil {
			// String literals
			g.Printf("%s:\n", o.Name)
			g.Printf("  .asciz \"%s\"\n", string(global.Val.([]byte)))
		} else {
			// Global variable
			g.Printf("  .comm _%s, %d\n", o.Name, o.Type.Size)
		}
	}
}

func (g *CodeGenerator) GenCode() {
	for _, o := range g.objects {
		function, ok := o.Val.(*Function)
		if !ok {
			continue
		}

		// Prologue
		g.Printf("  .globl _%s\n", o.Name)
		g.Printf("_%s:\n", o.Name)
		g.Printf("  push %%rbp\n")
		g.Printf("  mov %%rsp, %%rbp\n")
		g.Printf("  sub $%d, %%rsp\n", function.StackSize)

		// Prepare function parameters
		for i, param := range function.Params {
			bits := 64
			if param.Type.Size == 1 {
				bits = 8
			}
			g.Printf("  mov %s, %d(%%rbp)\n", ArgRegisters(i, bits), param.Val.(*Local).Offset)
		}

		g.GenStmt(function.Body, o.Name)
		g.CheckDepth()

		// Epilogue
		g.Printf(".L.return.%s:\n", o.Name)
		g.Printf("  mov %%rbp, %%rsp\n")
		g.Printf("  pop %%rbp\n")
		g.Printf("  ret\n")
	}
}

func (g *CodeGenerator) CheckDepth() {
	if g.depth != 0 {
		panic(errors.New("expression not closed"))
	}

	return
}

// Push data in rax to stack
func (g *CodeGenerator) Push() {
	g.Printf("  push %%rax\n")
	g.depth++
}

// Pop data from stack to <arg> register
func (g *CodeGenerator) Pop(arg string) {
	g.Printf("  pop %s\n", arg)
	g.depth--
}

// Load value in memory pointed by rax to rax
func (g *CodeGenerator) Load(t *Type) {
	if t.Kind == TYArray || t.Kind == TYStruct || t.Kind == TYUnion {
		return
	}

	if t.Size == 1 {
		g.Printf("  movsbq (%%rax), %%rax\n")
	} else {
		g.Printf("  mov (%%rax), %%rax\n")
	}
}

// Store rax to memory pointed by stack top
func (g *CodeGenerator) Store(t *Type) {
	g.Pop("%rdi")

	if t.Kind == TYStruct || t.Kind == TYUnion {
		for i := 0; i < t.Size; i++ {
			g.Printf("  mov %d(%%rax), %%r8b\n", i)
			g.Printf("  mov %%r8b, %d(%%rdi)\n", i)
		}

		return
	}

	if t.Size == 1 {
		g.Printf("  mov %%al, (%%rdi)\n")
	} else {
		g.Printf("  mov %%rax, (%%rdi)\n")
	}
}

// GenAddr puts node's memory address to rax
// But if node is a deref expr, the addr effect will be cancelled out
func (g *CodeGenerator) GenAddr(node *Node, funcName string) {
	switch node.Kind {
	case NKVariable:
		if _, ok := node.Val.(*Object).Val.(*Local); ok {
			g.Printf("  lea %d(%%rbp), %%rax\n", node.Val.(*Object).Val.(*Local).Offset)
		} else if node.Val.(*Object).Val.(*Global).Val != nil {
			g.Printf("  lea %s(%%rip), %%rax\n", node.Val.(*Object).Name)
		} else {
			g.Printf("  lea _%s(%%rip), %%rax\n", node.Val.(*Object).Name)
		}
		return
	case NKDeRef:
		g.GenExpr(node.Val.(*Node), funcName)
		return
	case NKComma:
		binary := node.Val.(*BinaryExpr)
		g.GenExpr(binary.Lhs, funcName)
		g.GenAddr(binary.Rhs, funcName)
		return
	case NKMember:
		g.GenAddr(node.Val.(*StructMemberAccess).Struct, funcName)
		g.Printf("  add $%d, %%rax\n", node.Val.(*StructMemberAccess).Member.Offset)
		return
	}

	panic(errors.New("not a lvalue"))
}

func (g *CodeGenerator) GenStmt(node *Node, funcName string) {
	switch node.Kind {
	case NKIf:
		c := g.LabelCount()
		ifClause := node.Val.(*IfClause)
		g.GenExpr(ifClause.Cond, funcName)
		g.Printf("  cmp $0, %%rax\n")
		g.Printf("  je .L.else.%d\n", c)
		g.GenStmt(ifClause.Then, funcName)
		g.Printf("  jmp .L.end.%d\n", c)
		g.Printf(".L.else.%d:\n", c)
		if ifClause.Else != nil {
			g.GenStmt(ifClause.Else, funcName)
		}
		g.Printf(".L.end.%d:\n", c)
		return
	case NKFor:
		c := g.LabelCount()
		forClause := node.Val.(*ForClause)
		if forClause.Init != nil {
			g.GenStmt(forClause.Init, funcName)
		}

		g.Printf(".L.begin.%d:\n", c)
		if forClause.Cond != nil {
			g.GenExpr(forClause.Cond, funcName)
			g.Printf("  cmp $0, %%rax\n")
			g.Printf("  je  .L.end.%d\n", c)
		}
		g.GenStmt(forClause.Body, funcName)
		if forClause.Increment != nil {
			g.GenExpr(forClause.Increment, funcName)
		}
		g.Printf("  jmp .L.begin.%d\n", c)
		g.Printf(".L.end.%d:\n", c)
		return
	case NKBlock:
		for _, n := range node.Val.([]*Node) {
			g.GenStmt(n, funcName)
		}
		return
	case NKReturn:
		g.GenExpr(node.Val.(*Node), funcName)
		g.Printf("  jmp .L.return.%s\n", funcName)
		return
	case NKExprStmt:
		g.GenExpr(node.Val.(*Node), funcName)
		return
	}

	panic(errors.New("invalid statement"))
}

func (g *CodeGenerator) GenExpr(node *Node, funcName string) {
	switch node.Kind {
	case NKNum:
		g.Printf("  mov $%d, %%rax\n", node.Val)
		return
	case NKNeg:
		g.GenExpr(node.Val.(*Node), funcName)
		g.Printf("  neg %%rax\n")
		return
	case NKVariable, NKMember:
		g.GenAddr(node, funcName)
		g.Load(node.Type)
		return
	case NKDeRef:
		g.GenExpr(node.Val.(*Node), funcName)
		g.Load(node.Type)
		return
	case NKAddr:
		g.GenAddr(node.Val.(*Node), funcName)
		return
	case NKAssign:
		binary := node.Val.(*BinaryExpr)
		g.GenAddr(binary.Lhs, funcName)
		g.Push()
		g.GenExpr(binary.Rhs, funcName)
		g.Store(node.Type)
		return
	case NKComma:
		binary := node.Val.(*BinaryExpr)
		g.GenExpr(binary.Lhs, funcName)
		g.GenExpr(binary.Rhs, funcName)
		return
	case NKStmtExpr:
		g.GenStmt(node.Val.(*Node), funcName)
		return
	case NKFuncCall:
		fc := node.Val.(*FuncCall)
		for _, arg := range fc.Args {
			g.GenExpr(arg, funcName)
			g.Push()
		}
		for i := len(fc.Args) - 1; i >= 0; i-- {
			g.Pop(ArgRegisters(i, 64))
		}

		g.Printf("  mov $0, %%rax\n")
		g.Printf("  call %s\n", "_"+fc.Name)
		return
	}

	binary := node.Val.(*BinaryExpr)
	g.GenExpr(binary.Rhs, funcName)
	g.Push()
	g.GenExpr(binary.Lhs, funcName)
	g.Pop("%rdi")

	switch node.Kind {
	case NKAdd:
		g.Printf("  add %%rdi, %%rax\n")
		return
	case NKSub:
		g.Printf("  sub %%rdi, %%rax\n")
		return
	case NKMul:
		g.Printf("  imul %%rdi, %%rax\n")
		return
	case NKDiv:
		g.Printf("  cqo\n")
		g.Printf("  idiv %%rdi\n")
		return
	case NKEq, NKNe, NKLt, NKLe:
		g.Printf("  cmp %%rdi, %%rax\n")
		if node.Kind == NKEq {
			g.Printf("  sete %%al\n")
		} else if node.Kind == NKNe {
			g.Printf("  setne %%al\n")
		} else if node.Kind == NKLt {
			g.Printf("  setl %%al\n")
		} else if node.Kind == NKLe {
			g.Printf("  setle %%al\n")
		}

		g.Printf("  movzb %%al, %%rax\n")
		return
	}

	panic(errors.New("invalid expression"))
}

func ArgRegisters(i, bits int) string {
	switch bits {
	case 8:
		return []string{"%dil", "%sil", "%dl", "%cl", "%r8b", "%r9b"}[i]
	case 64:
		return []string{"%rdi", "%rsi", "%rdx", "%rcx", "%r8", "%r9"}[i]
	default:
		return ""
	}
}
