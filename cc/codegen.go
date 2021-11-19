package cc

import "C"
import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const StackSize = 65536

type Codegen struct {
	objects    []*Object
	depth      int
	labelCount int
	writer     io.Writer
}

func (c *Codegen) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(c.writer, format, a...)
}

func NewCodegen(w io.Writer, objects []*Object) *Codegen {
	return &Codegen{writer: w, objects: objects}
}

func (c *Codegen) Gen() (err error) {
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
	c.Printf("(module\n")
	c.GenData()
	c.GenCode()

	// TODO: It ought to be enough to everyone.
	c.Printf("  (memory $memory (export \"memory\") 2)\n")
	c.Printf("  (global $sp (mut i32) (i32.const %d))\n", StackSize)
	c.Printf("  (global $bp i32       (i32.const %d))\n", StackSize)
	c.Printf(")\n")
	return
}

func (c *Codegen) GenData() int {
	memoryOffset := 0
	for _, o := range c.objects {
		if o.Kind != ObjectKindGlobal {
			continue
		}

		if o.Global.Val != nil {
			// String literals
			c.Printf("  (data (i32.const %d) \"%s\")\n", memoryOffset, o.Global.Val.([]byte))
			o.Global.Offset = memoryOffset
			memoryOffset += len(o.Global.Val.([]byte))
		} else {
			// Global variables
			c.Printf("  (data (i32.const %d) \"%s\")\n", memoryOffset, strings.Repeat("\\00", o.Type.Size))
			o.Global.Offset = memoryOffset
			memoryOffset += o.Type.Size
		}
	}

	return memoryOffset
}

func (c *Codegen) GenCode() {
	for _, o := range c.objects {
		if o.Kind != ObjectKindFunction || o.Function.IsDefinition {
			continue
		}

		c.Printf("  (func $%s (export \"%s\")", o.Name, o.Name)
		for _, param := range o.Function.Params {
			c.Printf(" (param $%s %s)", param.Name, param.Type.WasmType())
		}
		c.Printf(" (result i32)\n")

		// Prologue
		c.Printf("    global.get $sp\n")
		c.Printf("    i32.const %d\n", o.Function.StackSize)
		c.Printf("    i32.sub\n")
		c.Printf("    global.set $sp\n")

		c.GenStmt(o.Function.Body)
		c.Printf("    return\n")
		c.Printf("  )\n")
	}
}

func (c *Codegen) GenStmt(node *Node) {
	switch node.Kind {
	case NKBlock, NKStmtsExpr:
		for _, n := range node.Block.Stmts {
			c.GenStmt(n)
		}
		return
	case NKReturn:
		c.GenExpr(node.Unary.Expr)
		c.Printf("    return\n")
		return
	case NKExprStmt:
		c.GenExpr(node.Unary.Expr)
		return
	}

	panic(errors.New("invalid statement"))
}

func (c *Codegen) GenExpr(node *Node) {
	switch node.Kind {
	case NKNum:
		c.Printf("    %s.const %d\n", node.Type.WasmType(), node.Num.Val)
		return
	case NKNeg:
		c.Printf("    %s.const 0\n", node.Type.WasmType())
		c.GenExpr(node.Unary.Expr)
		c.Printf("    %s.sub\n", node.Type.WasmType())
		return
	case NKVariable:
		c.GenAddr(node)
		c.Printf("    %s.load\n", node.Type.WasmType())
		return
	case NKDeRef:
		c.GenExpr(node.Unary.Expr)
		c.Printf("    %s.load\n", node.Type.WasmType())
		return
	case NKAddr:
		c.GenAddr(node.Unary.Expr)
		return
	case NKAssign:
		c.GenAddr(node.Binary.Lhs)
		c.GenExpr(node.Binary.Rhs)
		c.Printf("    %s.store\n", node.Type.WasmType())
		c.GenExpr(node.Binary.Lhs)
		return
	case NKComma:
		c.GenExpr(node.Binary.Lhs)
		c.GenExpr(node.Binary.Rhs)
		return
	case NKStmtsExpr:
		c.GenStmt(node)
		return
	}

	c.GenExpr(node.Binary.Lhs)
	c.GenExpr(node.Binary.Rhs)

	switch node.Kind {
	case NKAdd:
		c.Printf("    %s.add\n", node.Type.WasmType())
		return
	case NKSub:
		c.Printf("    %s.sub\n", node.Type.WasmType())
		return
	case NKMul:
		c.Printf("    %s.mul\n", node.Type.WasmType())
		return
	case NKDiv:
		c.Printf("    %s.div_s\n", node.Type.WasmType())
		return
	case NKEq:
		c.Printf("    %s.eq\n", node.Type.WasmType())
		return
	case NKNe:
		c.Printf("    %s.ne\n", node.Type.WasmType())
		return
	case NKLt:
		c.Printf("    %s.lt_s\n", node.Type.WasmType())
		return
	case NKLe:
		c.Printf("    %s.le_s\n", node.Type.WasmType())
		return
	}

	panic(errors.New("invalid expression"))
}

func (c *Codegen) GenAddr(node *Node) {
	switch node.Kind {
	case NKVariable:
		switch node.Variable.Object.Kind {
		case ObjectKindLocal:
			c.Printf("    global.get $sp\n")
			c.Printf("    i32.const %d\n", node.Variable.Object.Local.Offset)
			c.Printf("    i32.add\n")
		case ObjectKindGlobal:
			c.Printf("    i32.const %d\n", node.Variable.Object.Global.Offset)
		default:
			panic(errors.New("not a lvalue"))
		}
		return
	case NKDeRef:
		c.GenExpr(node.Unary.Expr)
		return
	}

	panic(errors.New("not a lvalue"))
}
