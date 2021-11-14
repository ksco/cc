package tests

import "testing"

func TestArithmetic(t *testing.T) {
	a := Assert{t: t}
	a.Eval(int32(0), "int main() { return 0; }")
	a.Eval(int32(42), "int main() { return 42; }")
	a.Eval(int32(21), "int main() { return 5+20-4; }")
	a.Eval(int32(41), "int main() { return  12 + 34 - 5 ; }")
	a.Eval(int32(47), "int main() { return 5+6*7; }")
	a.Eval(int32(15), "int main() { return 5*(9-6); }")
	a.Eval(int32(4), "int main() { return (3+5)/2; }")
	a.Eval(int32(10), "int main() { return -10+20; }")
	a.Eval(int32(10), "int main() { return - -10; }")
	a.Eval(int32(10), "int main() { return - - +10; }")

	a.Eval(int32(0), "int main() { return 0==1; }")
	a.Eval(int32(1), "int main() { return 42==42; }")
	a.Eval(int32(1), "int main() { return 0!=1; }")
	a.Eval(int32(0), "int main() { return 42!=42; }")

	a.Eval(int32(1), "int main() { return 0<1; }")
	a.Eval(int32(0), "int main() { return 1<1; }")
	a.Eval(int32(0), "int main() { return 2<1; }")
	a.Eval(int32(1), "int main() { return 0<=1; }")
	a.Eval(int32(1), "int main() { return 1<=1; }")
	a.Eval(int32(0), "int main() { return 2<=1; }")

	a.Eval(int32(1), "int main() { return 1>0; }")
	a.Eval(int32(0), "int main() { return 1>1; }")
	a.Eval(int32(0), "int main() { return 1>2; }")
	a.Eval(int32(1), "int main() { return 1>=0; }")
	a.Eval(int32(1), "int main() { return 1>=1; }")
	a.Eval(int32(0), "int main() { return 1>=2; }")
}
