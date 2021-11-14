package cc

import "C"
import (
	"errors"
	"fmt"
	"io"
)

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
		for _, local := range function.Locals {
			c.Printf("    (local $%s %s)\n", local.Name, local.Type.WasmType())
		}
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
		c.Printf("    local.get $%s\n", node.Val.(*Object).Name)
		return
	case NKAssign:
		binary := node.Val.(*BinaryExpr)
		c.GenExpr(binary.Rhs)
		c.Printf("    local.set $%s\n", binary.Lhs.Val.(*Object).Name)
		c.Printf("    local.get $%s\n", binary.Lhs.Val.(*Object).Name)
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
