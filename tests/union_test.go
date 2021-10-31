package tests

import "testing"

func TestUnion(t *testing.T) {
	a := Assert{t: t}
	a.Eval(8, "int main() { union { int a; char b[6]; } x; return sizeof(x); }")
	a.Eval(3, "int main() { union { int a; char b[4]; } x; x.a = 515; return x.b[0]; }")
	a.Eval(2, "int main() { union { int a; char b[4]; } x; x.a = 515; return x.b[1]; }")
	a.Eval(0, "int main() { union { int a; char b[4]; } x; x.a = 515; return x.b[2]; }")
	a.Eval(0, "int main() { union { int a; char b[4]; } x; x.a = 515; return x.b[3]; }")

	a.Eval(3, "int main() { union t {char a; char b;} x; struct t *y = &x; x.a=3; return y->b; }")
	a.Eval(3, "int main() { union t {char a;} x; struct t *y = &x; y->a=3; return x.a; }")

	a.Eval(3, "int main() { union {int a,b;} x,y; x.a=3; y.a=5; y=x; return y.a; }")
	a.Eval(3, "int main() { union {struct {int a,b;} c;} x,y; x.c.b=3; y.c.b=5; y=x; return y.c.b; }")
}
