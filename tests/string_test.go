package tests

import "testing"

func TestString(t *testing.T) {
	a := Assert{t: t}
	a.Eval(int32(0), `int main() { return ""[0]; }`)
	a.Eval(int32(1), `int main() { return sizeof(""); }`)

	a.Eval(int32(97), `int main() { char* a = "abc"; return a[0]; }`)
	a.Eval(int32(97), `int main() { return "abc"[0]; }`)
	a.Eval(int32(98), `int main() { return "abc"[1]; }`)
	a.Eval(int32(99), `int main() { return "abc"[2]; }`)
	a.Eval(int32(0), `int main() { return "abc"[3]; }`)
	a.Eval(int32(4), `int main() { return sizeof("abc"); }`)

	a.Eval(int32(7), `int main() { return "\a"[0]; }`)
	a.Eval(int32(8), `int main() { return "\b"[0]; }`)
	a.Eval(int32(9), `int main() { return "\t"[0]; }`)
	a.Eval(int32(10), `int main() { return "\n"[0]; }`)
	a.Eval(int32(11), `int main() { return "\v"[0]; }`)
	a.Eval(int32(12), `int main() { return "\f"[0]; }`)
	a.Eval(int32(13), `int main() { return "\r"[0]; }`)
	a.Eval(int32(27), `int main() { return "\e"[0]; }`)

	a.Eval(int32(106), `int main() { return "\j"[0]; }`)
	a.Eval(int32(107), `int main() { return "\k"[0]; }`)
	a.Eval(int32(108), `int main() { return "\l"[0]; }`)

	a.Eval(int32(7), `int main() { return "\ax\ny"[0]; }`)
	a.Eval(int32(120), `int main() { return "\ax\ny"[1]; }`)
	a.Eval(int32(10), `int main() { return "\ax\ny"[2]; }`)
	a.Eval(int32(121), `int main() { return "\ax\ny"[3]; }`)

	a.Eval(int32(0), `int main() { return "\0"[0]; }`)
	a.Eval(int32(16), `int main() { return "\20"[0]; }`)
	a.Eval(int32(65), `int main() { return "\101"[0]; }`)
	a.Eval(int32(104), `int main() { return "\1500"[0]; }`)

	a.Eval(int32(0), `int main() { return "\x00"[0]; }`)
	a.Eval(int32(119), `int main() { return "\x77"[0]; }`)
	a.Eval(int32(-91), `int main() { return "\xA5"[0]; }`)
	a.Eval(int32(-1), `int main() { return "\x00ff"[0]; }`)

	a.Eval(int32(1), `int main() { char x=1; return x; }`)
	a.Eval(int32(1), `int main() { char x=1; char y=2; return x; }`)
	a.Eval(int32(2), `int main() { char x=1; char y=2; return y; }`)

	a.Eval(int32(1), `int main() { char x; return sizeof(x); }`)
	a.Eval(int32(10), `int main() { char x[10]; return sizeof(x); }`)

	// FIXME
	// a.Eval(int32(1), `int main() { return sub_char(7, 3, 3); } int sub_char(char a, char b, char c) { return a-b-c; }`)
}
