package tests

import "testing"

func TestString(t *testing.T) {
	a := Assert{t: t}
	a.Eval(0, `int main() { return ""[0]; }`)
	a.Eval(1, `int main() { return sizeof(""); }`)

	a.Eval(97, `int main() { return "abc"[0]; }`)
	a.Eval(98, `int main() { return "abc"[1]; }`)
	a.Eval(99, `int main() { return "abc"[2]; }`)
	a.Eval(0, `int main() { return "abc"[3]; }`)
	a.Eval(4, `int main() { return sizeof("abc"); }`)

	a.Eval(7, `int main() { return "\a"[0]; }`)
	a.Eval(8, `int main() { return "\b"[0]; }`)
	a.Eval(9, `int main() { return "\t"[0]; }`)
	a.Eval(10, `int main() { return "\n"[0]; }`)
	a.Eval(11, `int main() { return "\v"[0]; }`)
	a.Eval(12, `int main() { return "\f"[0]; }`)
	a.Eval(13, `int main() { return "\r"[0]; }`)
	a.Eval(27, `int main() { return "\e"[0]; }`)

	a.Eval(106, `int main() { return "\j"[0]; }`)
	a.Eval(107, `int main() { return "\k"[0]; }`)
	a.Eval(108, `int main() { return "\l"[0]; }`)

	a.Eval(7, `int main() { return "\ax\ny"[0]; }`)
	a.Eval(120, `int main() { return "\ax\ny"[1]; }`)
	a.Eval(10, `int main() { return "\ax\ny"[2]; }`)
	a.Eval(121, `int main() { return "\ax\ny"[3]; }`)

	a.Eval(0, `int main() { return "\0"[0]; }`)
	a.Eval(16, `int main() { return "\20"[0]; }`)
	a.Eval(65, `int main() { return "\101"[0]; }`)
	a.Eval(104, `int main() { return "\1500"[0]; }`)

	a.Eval(0, `int main() { return "\x00"[0]; }`)
	a.Eval(119, `int main() { return "\x77"[0]; }`)
	a.Eval(165, `int main() { return "\xA5"[0]; }`)
	a.Eval(255, `int main() { return "\x00ff"[0]; }`)

	a.Eval(1, `int main() { char x=1; return x; }`)
	a.Eval(1, `int main() { char x=1; char y=2; return x; }`)
	a.Eval(2, `int main() { char x=1; char y=2; return y; }`)

	a.Eval(1, `int main() { char x; return sizeof(x); }`)
	a.Eval(10, `int main() { char x[10]; return sizeof(x); }`)
	a.Eval(1, `int main() { return sub_char(7, 3, 3); } int sub_char(char a, char b, char c) { return a-b-c; }`)

}
