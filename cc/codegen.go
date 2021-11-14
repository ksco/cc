package cc

import "C"
import (
	"errors"
	"fmt"
	"io"
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
	c.GenCode()

	// TODO: It ought to be enough to everyone.
	c.Printf("  (memory $memory (export \"memory\") 2)\n")
	c.Printf("  (global $sp (mut i32) (i32.const %d))\n", StackSize)
	c.Printf("  (global $bp i32       (i32.const %d))\n", StackSize)
	c.Printf(")\n")
	return
}

func (c *Codegen) GenCode() {
	for _, o := range c.objects {
		function, ok := o.Val.(*Function)
		if !ok || function.IsDefinition {
			continue
		}

		c.Printf("  (func $%s (export \"%s\")", o.Name, o.Name)
		for _, param := range function.Params {
			c.Printf(" (param $%s %s)", param.Name, param.Type.WasmType())
		}
		c.Printf(" (result i32)\n")

		// Prologue
		c.Printf("    global.get $sp\n")
		c.Printf("    i32.const %d\n", function.StackSize)
		c.Printf("    i32.sub\n")
		c.Printf("    global.set $sp\n")

		c.GenStmt(function.Body)
		c.Printf("    return\n")
		c.Printf("  )\n")
	}
}

func (c *Codegen) GenStmt(node *Node) {
	switch node.Kind {
	case NKBlock:
		for _, n := range node.Val.([]*Node) {
			c.GenStmt(n)
		}
		return
	case NKReturn:
		c.GenExpr(node.Val.(*Node))
		c.Printf("    return\n")
		return
	case NKExprStmt:
		c.GenExpr(node.Val.(*Node))
		return
	}

	panic(errors.New("invalid statement"))
}

func (c *Codegen) GenExpr(node *Node) {
	switch node.Kind {
	case NKNum:
		c.Printf("    %s.const %d\n", node.Type.WasmType(), node.Val)
		return
	case NKNeg:
		c.Printf("    %s.const 0\n", node.Type.WasmType())
		c.GenExpr(node.Val.(*Node))
		c.Printf("    %s.sub\n", node.Type.WasmType())
		return
	case NKVariable:
		c.Printf("    global.get $sp\n")
		c.Printf("    %s.load offset=%d\n", node.Type.WasmType(), node.Val.(*Object).Val.(*Local).Offset)
		return
	case NKDeRef:
		c.GenExpr(node.Val.(*Node))
		c.Printf("    %s.load\n", node.Type.WasmType())
		return
	case NKAddr:
		c.GenAddr(node.Val.(*Node))
		return
	case NKAssign:
		binary := node.Val.(*BinaryExpr)
		c.Printf("    global.get $sp\n")
		c.GenExpr(binary.Rhs)
		c.Printf("    %s.store offset=%d\n", binary.Lhs.Type.WasmType(), binary.Lhs.Val.(*Object).Val.(*Local).Offset)
		c.Printf("    global.get $sp\n")
		c.Printf("    %s.load offset=%d\n", binary.Lhs.Type.WasmType(), binary.Lhs.Val.(*Object).Val.(*Local).Offset)
		return
	}

	binary := node.Val.(*BinaryExpr)
	c.GenExpr(binary.Lhs)
	c.GenExpr(binary.Rhs)

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
		c.Printf("    global.get $sp\n")
		c.Printf("    i32.const %d\n", node.Val.(*Object).Val.(*Local).Offset)
		c.Printf("    i32.add\n")
		return
	}

	panic(errors.New("not a lvalue"))
}
