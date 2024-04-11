package ennet

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"unicode"
)

func (l *Lexer) Dump() string {
	return "*lexer dump:\n" +
		fmt.Sprintf("  %+v\n", l.scanning) +
		fmt.Sprintf("  %+v\n", l.scanpos)
}

type Lexer struct {
	in *bufio.Reader

	// if len(scanning) == scanpos: next is read from in
	// else: next is from scannin[scanpos], then scanpos++
	scanning []Token
	scanpos  int
}

func NewLexer(in io.Reader) *Lexer {
	r := readerPool.Get().(*bufio.Reader)
	r.Reset(in)
	return &Lexer{
		in:       r,
		scanning: make([]Token, 0, 16),
	}
}

func (l *Lexer) Close() {
	readerPool.Put(l.in)
	*l = Lexer{}
}

func (l *Lexer) Next() Token {
	var tok Token
	if l.scanpos == len(l.scanning) {
		tok = l.scanNext()
		l.scanning = append(l.scanning, tok)
	} else { // l.scanpos < len(l.scanning)
		tok = l.scanning[l.scanpos]
	}
	l.scanpos++

	return tok
}

func (l *Lexer) Back() {
	if l.scanpos == 0 {
		return
	}
	l.scanpos--
}

func (l *Lexer) Transaction() LexerTx {
	return LexerTx{
		lexer: l,
		count: 0,
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
		text := []rune{}
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
					text = append(text, r)
				} else if sz == 0 {
					break
				} else {
					l.in.UnreadRune()
					break
				}
			} else {
				text = append(text, r)
			}
		}

		return Token{
			Type: QTEXT,
			Text: string(text),
		}

	case '{':
		text := []rune{}
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
					text = append(text, r)
				} else if sz == 0 {
					break
				} else {
					l.in.UnreadRune()
					break
				}
			} else {
				text = append(text, r)
			}
		}

		return Token{
			Type: TEXT,
			Text: string(text),
		}

	default:
		if unicode.IsSpace(r) {
			//fmt.Fprintf(os.Stderr, "skipping ... %q\n", string(r))
			l.skipSpace(r)
			return l.scanNext()
		}

		// STRING
		id := []rune{r}
		for {
			r, sz, err = l.in.ReadRune()
			if sz == 0 {
				return Token{
					Type: STRING,
					Text: string(id),
				}
			} else if err != nil {
				return Token{
					Type: ERR,
					Text: err.Error(),
				}
			}

			if isSTRING(r) {
				id = append(id, r)
			} else {
				l.in.UnreadRune()
				return Token{
					Type: STRING,
					Text: string(id),
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

var readerPool = sync.Pool{
	New: func() any {
		return bufio.NewReader(nil)
	},
}
