package ennet

import (
	"unicode"
)

type TokenType uint8

const (
	EOF TokenType = iota
	ERR

	CHILD
	SIBLING
	CLIMBUP
	MULT
	GROUPBEGIN
	GROUPEND

	ID
	CLASS
	ATTRBEGIN
	ATTREND
	EQ

	STRING
	TEXT
	QTEXT
)

var tokType2String = map[TokenType]string{
	EOF:        "eof",
	ERR:        "error",
	CHILD:      ">",
	SIBLING:    "+",
	CLIMBUP:    "^",
	MULT:       "*",
	GROUPBEGIN: "(",
	GROUPEND:   ")",
	ID:         "#",
	CLASS:      ".",
	ATTRBEGIN:  "[",
	ATTREND:    "]",
	EQ:         "=",
	STRING:     "STRING",
	TEXT:       "TEXT",
	QTEXT:      "QTEXT",
}

func (t TokenType) String() string {
	if s, found := tokType2String[t]; found {
		return s
	}
	return "???"
}

type Token struct {
	Type TokenType
	Text string
	Pos  int
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
		return t.Type.String()
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
