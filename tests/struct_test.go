package tests

import "testing"

func TestStruct(t *testing.T) {
	a := Assert{t: t}
	a.Eval(1, "int main() { struct {int a; int b;} x; x.a=1; x.b=2; return x.a; }")
	a.Eval(2, "int main() { struct {int a; int b;} x; x.a=1; x.b=2; return x.b; }")
	a.Eval(1, "int main() { struct {char a; int b; char c;} x; x.a=1; x.b=2; x.c=3; return x.a; }")
	a.Eval(2, "int main() { struct {char a; int b; char c;} x; x.b=1; x.b=2; x.c=3; return x.b; }")
	a.Eval(3, "int main() { struct {char a; int b; char c;} x; x.a=1; x.b=2; x.c=3; return x.c; }")

	a.Eval(0, "int main() { struct {char a; char b;} x[3]; char *p=x; p[0]=0; return x[0].a; }")
	a.Eval(1, "int main() { struct {char a; char b;} x[3]; char *p=x; p[1]=1; return x[0].b; }")
	a.Eval(2, "int main() { struct {char a; char b;} x[3]; char *p=x; p[2]=2; return x[1].a; }")
	a.Eval(3, "int main() { struct {char a; char b;} x[3]; char *p=x; p[3]=3; return x[1].b; }")

	a.Eval(6, "int main() { struct {char a[3]; char b[5];} x; char *p=&x; x.a[0]=6; return p[0]; }")
	a.Eval(7, "int main() { struct {char a[3]; char b[5];} x; char *p=&x; x.b[0]=7; return p[3]; }")

	a.Eval(6, "int main() { struct { struct { char b; } a; } x; x.a.b=6; return x.a.b; }")

	a.Eval(8, "int main() { struct {int a;} x; return sizeof(x); }")
	a.Eval(16, "int main() { struct {int a; int b;} x; return sizeof(x); }")
	a.Eval(16, "int main() { struct {int a, b;} x; return sizeof(x); }")
	a.Eval(24, "int main() { struct {int a[3];} x; return sizeof(x); }")
	a.Eval(32, "int main() { struct {int a;} x[4]; return sizeof(x); }")
	a.Eval(48, "int main() { struct {int a[3];} x[2]; return sizeof(x); }")
	a.Eval(2, "int main() { struct {char a; char b;} x; return sizeof(x); }")
	a.Eval(0, "int main() { struct {} x; return sizeof(x); }")
	a.Eval(16, "int main() { struct {char a; int b;} x; return sizeof(x); }")
	a.Eval(16, "int main() { struct {int a; char b;} x; return sizeof(x); }")
}
