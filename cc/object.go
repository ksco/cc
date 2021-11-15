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

type Object struct {
	Name string
	Type *Type

	// One of *Local, *Global or *Function
	Val interface{}
}

func (o *Object) AlignLocals() *Object {
	var (
		f  *Function
		ok bool
	)
	if f, ok = o.Val.(*Function); !ok {
		return o
	}
	offset := 0
	for _, l := range f.Locals {
		offset += l.Type.Size
		offset = alignTo(offset, l.Type.Align)
		l.Val = &Local{Offset: offset}
	}
	f.StackSize = alignTo(offset, 16)
	return o
}
