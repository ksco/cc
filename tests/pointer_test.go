package tests

import "testing"

func TestPointer(t *testing.T) {
	a := Assert{t: t}
	a.Eval(3, "int main() { int x[2]; int *y=&x; *y=3; return *x; }")
	a.Eval(3, "int main() { int x[3]; *x=3; *(x+1)=4; *(x+2)=5; return *x; }")
	a.Eval(4, "int main() { int x[3]; *x=3; *(x+1)=4; *(x+2)=5; return *(x+1); }")
	a.Eval(5, "int main() { int x[3]; *x=3; *(x+1)=4; *(x+2)=5; return *(x+2); }")

	a.Eval(0, "int main() { int x[2][3]; int *y=x; *y=0; return **x; }")
	a.Eval(1, "int main() { int x[2][3]; int *y=x; *(y+1)=1; return *(*x+1); }")
	a.Eval(2, "int main() { int x[2][3]; int *y=x; *(y+2)=2; return *(*x+2); }")
	a.Eval(3, "int main() { int x[2][3]; int *y=x; *(y+3)=3; return **(x+1); }")
	a.Eval(4, "int main() { int x[2][3]; int *y=x; *(y+4)=4; return *(*(x+1)+1); }")
	a.Eval(5, "int main() { int x[2][3]; int *y=x; *(y+5)=5; return *(*(x+1)+2); }")

	a.Eval(3, "int main() { int x[3]; *x=3; x[1]=4; x[2]=5; return *x; }")
	a.Eval(4, "int main() { int x[3]; *x=3; x[1]=4; x[2]=5; return *(x+1); }")
	a.Eval(5, "int main() { int x[3]; *x=3; x[1]=4; x[2]=5; return *(x+2); }")
	a.Eval(5, "int main() { int x[3]; *x=3; x[1]=4; 2[x]=5; return *(x+2); }")

	a.Eval(0, "int main() { int x[2][3]; int *y=x; y[0]=0; return x[0][0]; }")
	a.Eval(1, "int main() { int x[2][3]; int *y=x; y[1]=1; return x[0][1]; }")
	a.Eval(2, "int main() { int x[2][3]; int *y=x; y[2]=2; return x[0][2]; }")
	a.Eval(3, "int main() { int x[2][3]; int *y=x; y[3]=3; return x[1][0]; }")
	a.Eval(4, "int main() { int x[2][3]; int *y=x; y[4]=4; return x[1][1]; }")
	a.Eval(5, "int main() { int x[2][3]; int *y=x; y[5]=5; return x[1][2]; }")
}
