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

	a.Eval(4, "int main() { struct {int a;} x; return sizeof(x); }")
	a.Eval(8, "int main() { struct {int a; int b;} x; return sizeof(x); }")
	a.Eval(8, "int main() { struct {int a, b;} x; return sizeof(x); }")
	a.Eval(12, "int main() { struct {int a[3];} x; return sizeof(x); }")
	a.Eval(16, "int main() { struct {int a;} x[4]; return sizeof(x); }")
	a.Eval(24, "int main() { struct {int a[3];} x[2]; return sizeof(x); }")
	a.Eval(2, "int main() { struct {char a; char b;} x; return sizeof(x); }")
	a.Eval(0, "int main() { struct {} x; return sizeof(x); }")
	a.Eval(8, "int main() { struct {char a; int b;} x; return sizeof(x); }")
	a.Eval(8, "int main() { struct {int a; char b;} x; return sizeof(x); }")
	a.Eval(8, "int main() { struct t {int a; int b;} x; struct t y; return sizeof(y); }")
	a.Eval(8, "int main() { struct t {int a; int b;}; struct t y; return sizeof(y); }")
	a.Eval(2, "int main() { struct t {char a[2];}; { struct t {char a[4];}; } struct t y; return sizeof(y); }")
	a.Eval(3, "int main() { struct t {int x;}; int t=1; struct t y; y.x=2; return t+y.x; }")

	a.Eval(3, "int main() { struct t {char a;} x; struct t *y = &x; x.a=3; return y->a; }")
	a.Eval(3, "int main() { struct t {char a;} x; struct t *y = &x; y->a=3; return x.a; }")

	a.Eval(3, "int main() { struct {int a,b;} x,y; x.a=3; y=x; return y.a; }")
	a.Eval(7, "int main() { struct t {int a,b;}; struct t x; x.a=7; struct t y; struct t *z=&y; *z=x; return y.a; }")
	a.Eval(7, "int main() { struct t {int a,b;}; struct t x; x.a=7; struct t y, *p=&x, *q=&y; *q=*p; return y.a; }")
	a.Eval(5, "int main() { struct t {char a, b;} x, y; x.a=5; y=x; return y.a; }")
	a.Eval(3, "int main() { struct {int a,b;} x,y; x.a=3; y=x; return y.a; }")
	a.Eval(7, "int main() { struct t {int a,b;}; struct t x; x.a=7; struct t y; struct t *z=&y; *z=x; return y.a; }")
	a.Eval(7, "int main() { struct t {int a,b;}; struct t x; x.a=7; struct t y, *p=&x, *q=&y; *q=*p; return y.a; }")
	a.Eval(5, "int main() { struct t {char a, b;} x, y; x.a=5; y=x; return y.a; }")

	a.Eval(16, "int main() { struct {char a; long b;} x; return sizeof x; }")
	a.Eval(4, "int main() { struct {char a; short b;} x; return sizeof x; }")
}
