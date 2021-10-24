package cc

import "math"

type TypeKind int

const (
	TYChar TypeKind = iota
	TYInt
	TYPtr
	TYFunc
	TYArray
	TYStruct
)

type Type struct {
	Kind  TypeKind
	Base  *Type
	Size  int
	Align int
	Val   interface{}
}

func (t *Type) IsInteger() bool {
	return t.Kind == TYInt || t.Kind == TYChar
}

var (
	IntType  = NewType(TYInt, nil, nil)
	CharType = NewType(TYChar, nil, nil)
)

func NewType(k TypeKind, base *Type, val interface{}) *Type {
	size, align := 1, 1
	switch k {
	case TYChar:
	case TYInt, TYPtr:
		size = 8
		align = 8
	case TYArray:
		size = base.Size * val.(int)
		align = base.Align
	case TYStruct:
		offset := 0
		for _, m := range val.([]*StructMember) {
			offset = alignTo(offset, m.Type.Align)
			m.Offset = offset
			offset += m.Type.Size
			align = int(math.Max(float64(align), float64(m.Type.Align)))
		}
		size = alignTo(offset, align)
	}
	return &Type{
		Kind:  k,
		Base:  base,
		Size:  size,
		Val:   val,
		Align: align,
	}
}
