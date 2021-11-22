package cc

type Local struct {
	Offset int
}

type Function struct {
	Body         *Node
	Params       []*Object
	Locals       []*Object
	IsDefinition bool
	StackSize    int
}

type Global struct {
	Offset int
	Val    interface{}
}

type ObjectKind int

const (
	OKLocal ObjectKind = iota
	OKGlobal
	OKStringLiteral
	OKFunction
)

type Object struct {
	Name string
	Kind ObjectKind
	Type *Type

	// Only one of the following fields will be set.
	Local    *Local
	Global   *Global
	Function *Function
}

func (o *Object) AlignLocals() *Object {
	if o.Kind != OKFunction {
		return o
	}

	offset := 0
	for _, l := range o.Function.Locals {
		offset += l.Type.Size
		offset = alignTo(offset, l.Type.Align)
		l.Local.Offset = offset
	}
	o.Function.StackSize = alignTo(offset, 16)
	return o
}
