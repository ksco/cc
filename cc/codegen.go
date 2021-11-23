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
	blockCount int
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
		if o.Kind == OKGlobal {
			c.Printf("  (data (i32.const %d) \"%s\")\n", memoryOffset, strings.Repeat("\\00", o.Type.Size))
			o.Global.Offset = memoryOffset
			memoryOffset += o.Type.Size
		} else if o.Kind == OKStringLiteral {
			c.Printf("  (data (i32.const %d) \"%s\\00\")\n", memoryOffset, o.Global.Val.([]byte))
			o.Global.Offset = memoryOffset
			memoryOffset += len(o.Global.Val.([]byte)) + 1
		}
	}

	return memoryOffset
}

func (c *Codegen) GenCode() {
	for _, o := range c.objects {
		if o.Kind != OKFunction || o.Function.IsDefinition {
			continue
		}

		c.Printf("  (func $%s (export \"%s\")", o.Name, o.Name)
		for _, param := range o.Function.Params {
			c.Printf(" (param $%s %s)", param.Name, param.Type.WasmType())
		}
		c.Printf(" (result i32)\n")
		c.Printf("    (local $result i32)\n")

		// Prologue
		c.Printf("    global.get $sp\n")
		c.Printf("    i32.const %d\n", o.Function.StackSize)
		c.Printf("    i32.sub\n")
		c.Printf("    global.set $sp\n")
		c.Printf("    block $ENTRY\n")

		c.GenStmt(o.Function.Body)

		// // Epilogue
		c.Printf("    end\n")
		c.Printf("    global.get $sp\n")
		c.Printf("    i32.const %d\n", o.Function.StackSize)
		c.Printf("    i32.add\n")
		c.Printf("    global.set $sp\n")
		c.Printf("    local.get $result\n")
		c.Printf("    return\n")
		c.Printf("  )\n")
	}
}

func (c *Codegen) GenStmt(node *Node) {
	switch node.Kind {
	case NKIf:
		blockFalseName := c.NextBlockName()
		blockTrueName := c.NextBlockName()
		c.Printf("    block %s\n", blockFalseName)
		c.Printf("    block %s\n", blockTrueName)
		c.GenExpr(node.IfClause.Cond)
		c.Printf("    i32.eqz\n")
		c.Printf("    br_if %s\n", blockTrueName)
		c.GenStmt(node.IfClause.Then)
		c.Printf("    br %s\n", blockFalseName)
		c.Printf("    end\n")
		if node.IfClause.Else != nil {
			c.GenStmt(node.IfClause.Else)
		}
		c.Printf("    br %s\n", blockFalseName)
		c.Printf("    end\n")
		return
	case NKBlock, NKStmtsExpr:
		for _, n := range node.Block.Stmts {
			c.GenStmt(n)
		}
		return
	case NKReturn:
		c.GenExpr(node.Unary.Expr)

		c.Printf("    local.set $result\n")
		c.Printf("    br $ENTRY\n")
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
		c.Printf("    %s.%s\n", node.Type.WasmType(), node.Type.WasmLoad())
		return
	case NKStringLiteral:
		c.GenAddr(node)
		return
	case NKDeRef:
		c.GenExpr(node.Unary.Expr)
		c.Printf("    %s.%s\n", node.Type.WasmType(), node.Type.WasmLoad())
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
	case NKVariable, NKStringLiteral:
		switch node.Variable.Object.Kind {
		case OKLocal:
			c.Printf("    global.get $sp\n")
			c.Printf("    i32.const %d\n", node.Variable.Object.Local.Offset)
			c.Printf("    i32.add\n")
		case OKGlobal, OKStringLiteral:
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

func (c *Codegen) NextBlockName() string {
	result := fmt.Sprintf("$B%d", c.blockCount)
	c.blockCount++
	return result
}
