package cc

import (
	"bufio"
	"fmt"
	"strings"
)

type TokenKind int

const (
	TKPunctuator TokenKind = iota
	TKIdentifier
	TKKeyword
	TKString
	TKNumber
	TKEof
	TKUnknown
)

type Pos struct {
	Col int
	Row int
}

func NewPos() Pos {
	return Pos{Col: 0, Row: 1}
}

type String struct {
	Type *Type
	Val  []byte
}

type Token struct {
	Kind   TokenKind
	Lexeme string
	Val    interface{}
	Pos    Pos
	Source []rune
}

func NewToken(kind TokenKind, lexeme string, pos Pos, val interface{}, source []rune) *Token {
	return &Token{
		Kind:   kind,
		Lexeme: lexeme,
		Val:    val,
		Pos:    pos,
		Source: source,
	}
}

func (t *Token) Equal(kind TokenKind, lexeme string) bool {
	return t.Kind == kind && t.Lexeme == lexeme
}

func (t *Token) Errorf(format string, a ...interface{}) error {
	s := fmt.Sprintf(format, a...)
	if t.Kind == TKEof {
		return fmt.Errorf("unexpected EOF, %s", s)
	}
	line := getLine(t.Source, t.Pos.Row)
	return fmt.Errorf(
		"[%d:%d] error occurred:\n%s\n%s%s^ %s\n",
		t.Pos.Row, t.Pos.Col, line,
		strings.Repeat(" ", t.Pos.Col),
		strings.Repeat("~", len(t.Lexeme)-1),
		s,
	)
}

func getLine(s []rune, n int) string {
	scanner := bufio.NewScanner(strings.NewReader(string(s)))
	for scanner.Scan() {
		n -= 1
		if n == 0 {
			return scanner.Text()
		}
	}

	return ""
}
