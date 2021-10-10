package cc

type TypeKind int

const (
	TypeKindChar TypeKind = iota
	TypeKindInt
	TypeKindPtr
	TypeKindFunc
	TypeKindArray
)

type Type struct {
	Kind TypeKind
	Base *Type
	Size int
	Val  interface{}
}

func (t *Type) IsInteger() bool {
	return t.Kind == TypeKindInt || t.Kind == TypeKindChar
}

var (
	IntType  = NewType(TypeKindInt, nil, nil)
	CharType = NewType(TypeKindChar, nil, nil)
)

func NewType(k TypeKind, base *Type, val interface{}) *Type {
	size := 0
	switch k {
	case TypeKindChar:
		size = 1
	case TypeKindInt, TypeKindPtr:
		size = 8
	case TypeKindArray:
		size = base.Size * val.(int)
	}
	return &Type{
		Kind: k,
		Base: base,
		Size: size,
		Val:  val,
	}
}
