package ennet

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

func (l *Lexer) Dump() string {
	return "*lexer dump:\n" +
		fmt.Sprintf("  %+v\n", l.scanning) +
		fmt.Sprintf("  %+v\n", l.scanned)
}

type Lexer struct {
	in *bufio.Reader

	// in --(scanNext)-> scanning --pop(Next)push-> scanned
	//                      ^----push(Back)pop--------|
	scanning, scanned []Token
}

func NewLexer(in io.Reader) *Lexer {
	return &Lexer{
		in:       bufio.NewReader(in),
		scanning: make([]Token, 0),
		scanned:  make([]Token, 0),
	}
}

func (l *Lexer) Next() Token {
	var tok Token
	if len(l.scanning) == 0 {
		tok = l.scanNext()
	} else {
		tok = l.scanning[len(l.scanning)-1]
		l.scanning = l.scanning[:len(l.scanning)-1]
	}

	l.scanned = append(l.scanned, tok)
	return tok
}

func (l *Lexer) Back() {
	if len(l.scanned) == 0 {
		return
	}
	tok := l.scanned[len(l.scanned)-1]
	l.scanned = l.scanned[:len(l.scanned)-1]
	l.scanning = append(l.scanning, tok)
}

func (l *Lexer) Transaction() LexerTx {
	return LexerTx{
		lexer:  l,
		count:  0,
		parent: nil,
	}
}

func (l *Lexer) scanNext() Token {
	r, sz, err := l.in.ReadRune()
	if sz == 0 && err != nil {
		return Token{
			Type: EOF,
		}
	} else if err != nil && err != io.EOF {
		return Token{
			Type: ERR,
			Text: err.Error(),
		}
	}

	switch r {
	case '>':
		return Token{
			Type: CHILD,
		}

	case '+':
		return Token{
			Type: SIBLING,
		}

	case '^':
		return Token{
			Type: CLIMBUP,
		}

	case '*':
		return Token{
			Type: MULT,
		}

	case '(':
		return Token{
			Type: GROUPBEGIN,
		}

	case ')':
		return Token{
			Type: GROUPEND,
		}

	case '#':
		return Token{
			Type: ID,
		}

	case '.':
		return Token{
			Type: CLASS,
		}

	case '[':
		return Token{
			Type: ATTRBEGIN,
		}

	case ']':
		return Token{
			Type: ATTREND,
		}

	case '=':
		return Token{
			Type: EQ,
		}

	case '\'', '"':
		quot := r
		text := ""
		for {
			r, sz, err = l.in.ReadRune()
			if sz == 0 {
				return Token{
					Type: ERR,
					Text: "sudden EOF",
				}
			} else if err != nil {
				return Token{
					Type: ERR,
					Text: err.Error(),
				}
			}

			if r == quot {

				rr, sz, _ := l.in.ReadRune()
				if rr == quot {
					text += string(r)
				} else if sz == 0 {
					break
				} else {
					l.in.UnreadRune()
					break
				}
			} else {
				text += string(r)
			}
		}

		return Token{
			Type: QTEXT,
			Text: text,
		}

	case '{':
		text := ""
		for {
			r, sz, err = l.in.ReadRune()
			if sz == 0 {
				return Token{
					Type: ERR,
					Text: "sudden EOF",
				}
			} else if err != nil {
				return Token{
					Type: ERR,
					Text: err.Error(),
				}
			}

			if r == '}' {

				rr, sz, _ := l.in.ReadRune()
				if rr == '}' {
					text += string(r)
				} else if sz == 0 {
					break
				} else {
					l.in.UnreadRune()
					break
				}
			} else {
				text += string(r)
			}
		}

		return Token{
			Type: TEXT,
			Text: text,
		}

	default:
		if unicode.IsSpace(r) {
			//fmt.Fprintf(os.Stderr, "skipping ... %q\n", string(r))
			l.skipSpace(r)
			return l.scanNext()
		}

		// STRING
		id := string(r)
		for {
			r, sz, err = l.in.ReadRune()
			if sz == 0 {
				return Token{
					Type: STRING,
					Text: id,
				}
			} else if err != nil {
				return Token{
					Type: ERR,
					Text: err.Error(),
				}
			}

			if isSTRING(r) {
				id += string(r)
			} else {
				l.in.UnreadRune()
				return Token{
					Type: STRING,
					Text: id,
				}
			}
		}
	}
}

func (l *Lexer) skipSpace(initr rune) {
	if unicode.IsSpace(initr) {
		for {
			r, sz, _ := l.in.ReadRune()
			if sz == 0 {
				return
			}

			//fmt.Fprintf(os.Stderr, "    next ... %q\n", string(r))
			if !unicode.IsSpace(r) {
				l.in.UnreadRune()
				return
			}
		}
	}
}
