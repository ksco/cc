package cc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Scanner struct {
	source []rune
	code   []rune
	pos    Pos
}

func NewScanner(code []rune) *Scanner {
	return &Scanner{source: code, code: code, pos: NewPos()}
}

func (s *Scanner) Scan() ([]*Token, error) {
	tokens := make([]*Token, 0)
	for len(s.code) > 0 {
		if len(s.code) > 1 && string(s.code[:2]) == "//" {
			s.skip(2)
			for len(s.code) > 0 && s.code[0] != '\n' {
				s.skip(1)
			}
			continue
		}

		if len(s.code) > 1 && string(s.code[:2]) == "/*" {
			s.skip(2)
			for len(s.code) > 0 {
				if len(s.code) > 1 && string(s.code[:2]) == "*/" {
					s.skip(2)
					break
				}
				s.skip(1)
			}
			if len(s.code) == 0 {
				return nil, errors.New("unclosed block comment")
			}
			continue
		}

		if unicode.IsSpace(s.code[0]) {
			s.skip(1)
			continue
		}

		if p, pl := readPunctuator(s.code); pl > 0 {
			tokens = append(tokens, NewToken(TKPunctuator, p, s.pos, nil, s.source))
			s.skip(pl)
			continue
		}

		// Parse variable or keyword
		if isAlpha(s.code[0]) {
			name, l := readIdentifier(s.code)
			if isKeyword(name) {
				tokens = append(tokens, NewToken(TKKeyword, name, s.pos, name, s.source))
			} else {
				tokens = append(tokens, NewToken(TKIdentifier, name, s.pos, name, s.source))
			}

			s.skip(l)
			continue
		}

		if unicode.IsDigit(s.code[0]) {
			var (
				num int
				l   int
				err error
			)
			if num, l, err = parseInt(s.code); err != nil {
				return nil, err
			}
			tokens = append(tokens, NewToken(TKNumber, string(s.code[:l]), s.pos, num, s.source))
			s.skip(l)
			continue
		}

		if s.code[0] == '"' {
			var (
				r   []rune
				l   int
				err error
			)
			if r, l, err = readStringLiteral(s.code); err != nil {
				return nil, err
			}
			tokens = append(tokens, NewToken(TKString, string(s.code[:l]), s.pos, &String{
				Type: NewType(TYArray, CharType, len(r)+1),
				Val:  []byte(string(r)),
			}, s.source))
			s.skip(l)
			continue
		}

		return nil, NewToken(TKUnknown, string(s.code[0]), s.pos, nil, s.source).Errorf("invalid token")
	}

	tokens = append(tokens, NewToken(TKEof, "", s.pos, nil, s.source))
	return tokens, nil
}

func (s *Scanner) skip(n int) {
	if n == 1 && s.code[0] == '\n' {
		s.pos.Col = 0
		s.pos.Row += 1
	} else {
		s.pos.Col += n
	}
	s.code = s.code[n:]
}

func parseInt(s []rune) (num int, l int, err error) {
	for l < len(s) && unicode.IsDigit(s[l]) {
		l += 1
	}
	result, err := strconv.ParseUint(string(s[0:l]), 10, 64)
	num = int(result)
	return
}

func readPunctuator(s []rune) (string, int) {
	if len(s) >= 2 {
		p := string(s[:2])
		switch p {
		case "==", "!=", "<=", ">=", "->":
			return p, 2
		}
	}

	if strings.ContainsRune("+-*/(){}<>[],;=&.", s[0]) {
		return string(s[0]), 1
	}

	return "", 0
}

func isAlpha(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || r == '_'
}

func isAlphaNumeric(r rune) bool {
	return isAlpha(r) || ('0' <= r && r <= '9')
}

func readIdentifier(s []rune) (i string, l int) {
	for l < len(s) && isAlphaNumeric(s[l]) {
		l += 1
	}
	i = string(s[0:l])
	return
}

func hexToInt(ch rune) rune {
	if ch >= '0' && ch <= '9' {
		return ch - '0'
	} else if ch >= 'A' && ch <= 'F' {
		return ch - 'A' + 10
	} else if ch >= 'a' && ch <= 'f' {
		return ch - 'a' + 10
	}

	return 0
}

func isHex(r rune) bool {
	if (r >= '0' && r <= '9') ||
		(r >= 'a' && r <= 'f') ||
		(r >= 'A' && r <= 'F') {
		return true
	}
	return false

}

func readEscapedChar(s []rune) (rs []rune, l int, err error) {
	switch s[l] {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		c := s[l] - '0'
		l += 1
		if l < len(s) && '0' <= s[l] && s[l] <= '7' {
			c = (c << 3) + s[l] - '0'
			l += 1
			if l < len(s) && '0' <= s[l] && s[l] <= '7' {
				c = (c << 3) + s[l] - '0'
				l += 1
			}
		}

		rs = []rune(fmt.Sprintf("\\%02x", c))
		return
	case 'x':
		l += 1
		if !(l < len(s) && isHex(s[l])) {
			err = errors.New("invalid hex escape sequence")
			return
		}

		var c rune
		for ; isHex(s[l]); l++ {
			c = (c << 4) + hexToInt(s[l])
		}
		rs = []rune(fmt.Sprintf("\\%02x", c))
		return
	case 'a':
		rs = []rune("\\07")
		l = l + 1
		return
	case 'b':
		rs = []rune("\\08")
		l = l + 1
		return
	case 'f':
		rs = []rune("\\0c")
		l = l + 1
		return
	case 'r':
		rs = []rune("\\0d")
		l = l + 1
		return
	case 'v':
		rs = []rune("\\0b")
		l = l + 1
		return
	case 'e':
		rs = []rune("\\1b")
		l = l + 1
		return
	case 't', 'n':
		rs = []rune{'\\', s[l]}
		l = l + 1
		return
	default:
		// TODO: warning: invalid escape sequence
		rs = []rune{s[l]}
		l = l + 1
		return
	}
}

func readStringLiteral(s []rune) (rs []rune, l int, err error) {
	l = 1
	for l < len(s) && s[l] != '"' {
		if s[l] == '\n' || s[l] == '\000' {
			err = errors.New("unclosed string literal")
			return
		}
		if s[l] == '\\' {
			if l+1 >= len(s) {
				err = errors.New("unclosed string literal")
				return
			}
			var (
				c  []rune
				ll int
			)
			c, ll, err = readEscapedChar(s[l+1:])
			if err != nil {
				return
			}
			rs = append(rs, c...)
			l += ll + 1
			continue
		}

		rs = append(rs, s[l])
		l += 1
	}

	if l >= len(s) {
		err = errors.New("unclosed string literal")
		return
	}

	l += 1
	return
}

func isKeyword(n string) bool {
	switch n {
	case "return", "if", "else", "for", "while", "long", "int", "short", "char", "sizeof", "struct", "union":
		return true
	}
	return false
}
