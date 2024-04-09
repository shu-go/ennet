package ennet

import "unicode"

type TokenType string

const (
	EOF = TokenType("eof")
	ERR = TokenType("err")

	CHILD      = TokenType(">")
	SIBLING    = TokenType("+")
	CLIMBUP    = TokenType("^")
	MULT       = TokenType("*")
	GROUPBEGIN = TokenType("(")
	GROUPEND   = TokenType(")")

	ID        = TokenType("#")
	CLASS     = TokenType(".")
	ATTRBEGIN = TokenType("[")
	ATTREND   = TokenType("]")
	EQ        = TokenType("=")

	STRING = TokenType("string")
	TEXT   = TokenType(`{}`)
	QTEXT  = TokenType(`""`)
)

type Token struct {
	Type TokenType
	Text string
}

func (t Token) String() string {
	switch t.Type {
	case ERR:
		return t.Text
	case STRING:
		return "string(" + t.Text + ")"
	case TEXT:
		return "{" + t.Text + "}"
	case QTEXT:
		return `"` + t.Text + `"`
	default:
		return string(t.Type)
	}
}

func isSTRING(r rune) bool {
	switch r {
	case '>', '+', '^', '*', '(', ')', '#', '.', '[', ']', '=', '{', '}':
		return false
	default:
		return !unicode.IsSpace(r)
	}
}
