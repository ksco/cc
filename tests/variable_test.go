package tests

import "testing"

func TestVariable(t *testing.T) {
	a := Assert{t: t}
	a.Eval(3, "int main() { int a; a=3; return a; }")
	a.Eval(3, "int main() { int a=3; return a; }")
	a.Eval(8, "int main() { int a=3; int z=5; return a+z; }")

	a.Eval(1, "int main() { return 1; 2; 3; }")
	a.Eval(2, "int main() { 1; return 2; 3; }")
	a.Eval(3, "int main() { 1; 2; return 3; }")

	a.Eval(3, "int main() { int a=3; return a; }")
	a.Eval(8, "int main() { int a=3; int z=5; return a+z; }")
	a.Eval(6, "int main() { int a; int b; a=b=3; return a+b; }")
	a.Eval(3, "int main() { int foo=3; return foo; }")
	a.Eval(8, "int main() { int foo123=3; int bar=5; return foo123+bar; }")

	a.Eval(3, "int main() { int x=3; return *&x; }")
	a.Eval(3, "int main() { int x=3; int *y=&x; int **z=&y; return **z; }")
	a.Eval(5, "int main() { int x=3; int y=5; return *(&x+1); }")
	a.Eval(3, "int main() { int x=3; int y=5; return *(&y-1); }")
	a.Eval(5, "int main() { int x=3; int y=5; return *(&x-(-1)); }")
	a.Eval(5, "int main() { int x=3; int *y=&x; *y=5; return x; }")
	a.Eval(7, "int main() { int x=3; int y=5; *(&x+1)=7; return y; }")
	a.Eval(7, "int main() { int x=3; int y=5; *(&y-2+1)=7; return x; }")
	a.Eval(5, "int main() { int x=3; return (&x+2)-&x+3; }")
	a.Eval(8, "int main() { int x, y; x=3; y=5; return x+y; }")
	a.Eval(8, "int main() { int x=3, y=5; return x+y; }")

	a.Eval(8, "int main() { int x; return sizeof(x); }")
	a.Eval(8, "int main() { int x; return sizeof x; }")
	a.Eval(8, "int main() { int *x; return sizeof(x); }")
	a.Eval(32, "int main() { int x[4]; return sizeof(x); }")
	a.Eval(96, "int main() { int x[3][4]; return sizeof(x); }")
	a.Eval(32, "int main() { int x[3][4]; return sizeof(*x); }")
	a.Eval(8, "int main() { int x[3][4]; return sizeof(**x); }")
	a.Eval(9, "int main() { int x[3][4]; return sizeof(**x) + 1; }")
	a.Eval(9, "int main() { int x[3][4]; return sizeof **x + 1; }")
	a.Eval(8, "int main() { int x[3][4]; return sizeof(**x + 1); }")
	a.Eval(8, "int main() { int x=1; return sizeof(x=2); }")
	a.Eval(1, "int main() { int x=1; sizeof(x=2); return x; }")

	a.Eval(0, "int x; int main() { return x; }")
	a.Eval(3, "int x; int main() { x=3; return x; }")
	a.Eval(7, "int x; int y; int main() { x=3; y=4; return x+y; }")
	a.Eval(7, "int x, y; int main() { x=3; y=4; return x+y; }")
	a.Eval(0, "int x[4]; int main() { x[0]=0; x[1]=1; x[2]=2; x[3]=3; return x[0]; }")
	a.Eval(1, "int x[4]; int main() { x[0]=0; x[1]=1; x[2]=2; x[3]=3; return x[1]; }")
	a.Eval(2, "int x[4]; int main() { x[0]=0; x[1]=1; x[2]=2; x[3]=3; return x[2]; }")
	a.Eval(3, "int x[4]; int main() { x[0]=0; x[1]=1; x[2]=2; x[3]=3; return x[3]; }")

	a.Eval(8, "int x; int main() { return sizeof(x); }")
	a.Eval(32, "int x[4]; int main() { return sizeof(x); }")

	a.Eval(0, "int main() { return ({ 0; }); }")
	a.Eval(2, "int main() { return ({ 0; 1; 2; }); }")
	a.Eval(1, "int main() { ({ 0; return 1; 2; }); return 3; }")
	a.Eval(6, "int main() { return ({ 1; }) + ({ 2; }) + ({ 3; }); }")
	a.Eval(3, "int main() { return ({ int x=3; x; }); }")

	a.Eval(2, "int main() { /* return 1; */ return 2; }")
	a.Eval(2, "int main() { // return 1;\nreturn 2; }")

	a.Eval(2, "int main() { int x=2; { int x=3; } return x; }")
	a.Eval(2, "int main() { int x=2; { int x=3; } { int y=4; return x; }}")
	a.Eval(3, "int main() { int x=2; { x=3; } return x; }")
}
