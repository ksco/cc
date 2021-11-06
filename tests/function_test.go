package tests

import "testing"

func TestFunction(t *testing.T) {
	a := Assert{t: t}
	a.Eval(32, "int main() { return ret32(); } int ret32() { return 32; }")
	a.Eval(7, "int main() { return add2(3,4); } int add2(int x, int y) { return x+y; }")
	a.Eval(1, "int main() { return sub2(4,3); } int sub2(int x, int y) { return x-y; }")
	a.Eval(55, "int main() { return fib(9); } int fib(int x) { if (x<=1) return 1; return fib(x-1) + fib(x-2); }")
	a.Eval(21, "int main() { return add6(1,2,3,4,5,6); } int add6(int a, int b, int c, int d, int e, int f) {return a+b+c+d+e+f;}")
	a.Eval(1, "int main() { return sub2(4,3); } int sub2(long x, long y) { return x-y; }")
	a.Eval(1, "int main() { return sub2(4,3); } int sub2(short x, short y) { return x-y; }")

	a.Eval(1, "int ret1(); int main() { return ret1(); } int ret1() { return 1; }")
}
