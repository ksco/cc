package tests

import (
	"cc/cc"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type Assert struct {
	t *testing.T
}

func (a Assert) Eval(result interface{}, s string) {
	sb := new(strings.Builder)
	err := cc.Compile(sb, []rune(s))
	if err != nil {
		a.t.Errorf("Compile failed, error: %s, code: %s", err.Error(), s)
	}

	err = os.WriteFile("/tmp/a.s", []byte(sb.String()), 0644)
	if err != nil {
		a.t.Errorf("Save file failed, error: %s, code: %s", err.Error(), s)
	}

	if err := exec.Command(
		"sh",
		"-c",
		fmt.Sprintf(`clang -o /tmp/a /tmp/a.s && /tmp/a`),
	).Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			if ee.ExitCode() == result {
				return
			}
			a.t.Errorf("Result error, expected: %v, got %d, code: %s", result, ee.ExitCode(), s)
			return
		}
		a.t.Errorf("Run program failed, error: %s, code: %s", err.Error(), s)
	}
}
