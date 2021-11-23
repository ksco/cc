package tests

import "testing"

func TestControl(t *testing.T) {
	a := Assert{t: t}
	a.Eval(int32(3), "int main() { if (0) return 2; return 3; }")
	a.Eval(int32(3), "int main() { if (1-1) return 2; else return 3; }")
	a.Eval(int32(2), "int main() { if (1) return 2; else return 3; }")
	a.Eval(int32(2), "int x; int main() { if (2-1) x = 2; else x = 3; return x; }")

	a.Eval(55, "int main() { int i=0; int j=0; for (i=0; i<=10; i=i+1) j=i+j; return j; }")
	a.Eval(3, "int main() { for (;;) return 3; return 5; }")

	a.Eval(10, "int main() { int i=0; while(i<10) i=i+1; return i; }")

	a.Eval(3, "int main() { {1; {2;} return 3;} }")
	a.Eval(5, "int main() { ;;; return 5; }")

	a.Eval(10, "int main() { int i=0; while(i<10) i=i+1; return i; }")
	a.Eval(55, "int main() { int i=0; int j=0; while(i<=10) {j=i+j; i=i+1;} return j; }")

	a.Eval(3, "int main() { return (1,2,3); }")
	a.Eval(5, "int main() { int i=2, j=3; (i=5,j)=6; return i; }")
	a.Eval(6, "int main() { int i=2, j=3; (i=5,j)=6; return j; }")
}
