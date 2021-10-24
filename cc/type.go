package cc

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
	Kind TypeKind
	Base *Type
	Size int
	Val  interface{}
}

func (t *Type) IsInteger() bool {
	return t.Kind == TYInt || t.Kind == TYChar
}

var (
	IntType  = NewType(TYInt, nil, nil)
	CharType = NewType(TYChar, nil, nil)
)

func NewType(k TypeKind, base *Type, val interface{}) *Type {
	size := 0
	switch k {
	case TYChar:
		size = 1
	case TYInt, TYPtr:
		size = 8
	case TYArray:
		size = base.Size * val.(int)
	}
	return &Type{
		Kind: k,
		Base: base,
		Size: size,
		Val:  val,
	}
}
