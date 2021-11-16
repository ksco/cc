package cc

import (
	"math"
)

type TypeKind int

const (
	TYLong TypeKind = iota
	TYPtr
	TYInt
	TYShort
	TYChar
	TYFunc
	TYArray
	TYStruct
	TYUnion
	TYUnknown
)

type StructVal struct {
	Members []*StructMember
	Name    *Token
}

type Type struct {
	Kind  TypeKind
	Base  *Type
	Size  int
	Align int
	Val   interface{}
}

func (t *Type) IsInteger() bool {
	return t.Kind == TYLong || t.Kind == TYInt || t.Kind == TYShort || t.Kind == TYChar
}

func (t *Type) WasmType() string {
	switch t.Kind {
	case TYLong:
		return "i64"
	case TYInt, TYPtr:
		return "i32"
	case TYArray:
		return t.Base.WasmType()
	default:
		return "i32"
	}
}

// Resize recalculates size and align recursively
func (t *Type) Resize() {
	if t.Base != nil {
		t.Base.Resize()
	}

	n := NewType(t.Kind, t.Base, t.Val)
	t.Size = n.Size
	t.Align = n.Align
}

var (
	LongType  = NewType(TYLong, nil, nil)
	ShortType = NewType(TYShort, nil, nil)
	IntType   = NewType(TYInt, nil, nil)
	CharType  = NewType(TYChar, nil, nil)
)

func NewType(k TypeKind, base *Type, val interface{}) *Type {
	size, align := 1, 1
	switch k {
	case TYChar:
		size, align = 1, 1
	case TYShort:
		size, align = 2, 2
	case TYInt, TYPtr:
		size, align = 4, 4
	case TYLong:
		size, align = 8, 8
	case TYArray:
		size = base.Size * val.(int)
		align = base.Align
	case TYStruct:
		offset := 0
		for _, m := range val.(*StructVal).Members {
			offset = alignTo(offset, m.Type.Align)
			m.Offset = offset
			offset += m.Type.Size
			align = int(math.Max(float64(align), float64(m.Type.Align)))
		}
		size = alignTo(offset, align)
	case TYUnion:
		for _, m := range val.(*StructVal).Members {
			if m.Type.Size > size {
				size = m.Type.Size
			}
			if m.Type.Align > align {
				align = m.Type.Align
			}
		}
		size = alignTo(size, align)
	}
	return &Type{
		Kind:  k,
		Base:  base,
		Size:  size,
		Val:   val,
		Align: align,
	}
}
