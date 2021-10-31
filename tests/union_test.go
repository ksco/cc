package tests

import "testing"

func TestUnion(t *testing.T) {
	a := Assert{t: t}
	a.Eval(8, "int main() { union { int a; char b[6]; } x; return sizeof(x); }")
	a.Eval(3, "int main() { union { int a; char b[4]; } x; x.a = 515; return x.b[0]; }")
	a.Eval(2, "int main() { union { int a; char b[4]; } x; x.a = 515; return x.b[1]; }")
	a.Eval(0, "int main() { union { int a; char b[4]; } x; x.a = 515; return x.b[2]; }")
	a.Eval(0, "int main() { union { int a; char b[4]; } x; x.a = 515; return x.b[3]; }")
}
