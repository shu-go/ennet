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
	in  *bufio.Reader
	pos int

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
		pos:      1,
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

func (l *Lexer) scanNext() Token {
	startpos := l.pos

	r, sz, err := l.in.ReadRune()
	if sz == 0 && err != nil {
		return Token{
			Type: EOF,
			Pos:  l.pos,
		}
	} else if err != nil && err != io.EOF {
		return Token{
			Type: ERR,
			Text: err.Error(),
			Pos:  l.pos,
		}
	}
	l.pos++

	switch r {
	case '>':
		return Token{
			Type: CHILD,
			Pos:  startpos,
		}

	case '+':
		return Token{
			Type: SIBLING,
			Pos:  startpos,
		}

	case '^':
		return Token{
			Type: CLIMBUP,
			Pos:  startpos,
		}

	case '*':
		return Token{
			Type: MULT,
			Pos:  startpos,
		}

	case '(':
		return Token{
			Type: GROUPBEGIN,
			Pos:  startpos,
		}

	case ')':
		return Token{
			Type: GROUPEND,
			Pos:  startpos,
		}

	case '#':
		return Token{
			Type: ID,
			Pos:  startpos,
		}

	case '.':
		return Token{
			Type: CLASS,
			Pos:  startpos,
		}

	case '[':
		return Token{
			Type: ATTRBEGIN,
			Pos:  startpos,
		}

	case ']':
		return Token{
			Type: ATTREND,
			Pos:  startpos,
		}

	case '=':
		return Token{
			Type: EQ,
			Pos:  startpos,
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
					Pos:  l.pos,
				}
			} else if err != nil {
				return Token{
					Type: ERR,
					Text: err.Error(),
					Pos:  l.pos,
				}
			}
			l.pos++

			if r == quot {
				rr, sz, _ := l.in.ReadRune()
				if rr == quot {
					l.pos++
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
			Pos:  startpos,
		}

	case '{':
		text := []rune{}
		for {
			r, sz, err = l.in.ReadRune()
			if sz == 0 {
				return Token{
					Type: ERR,
					Text: "sudden EOF",
					Pos:  l.pos,
				}
			} else if err != nil {
				return Token{
					Type: ERR,
					Text: err.Error(),
					Pos:  l.pos,
				}
			}
			l.pos++

			if r == '}' {
				rr, sz, _ := l.in.ReadRune()
				if rr == '}' {
					l.pos++
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
			Pos:  startpos,
		}

	default:
		if unicode.IsSpace(r) {
			//fmt.Fprintf(os.Stderr, "skipping ... %q\n", string(r))
			l.pos += l.skipSpace(r)
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
					Pos:  startpos,
				}
			} else if err != nil {
				return Token{
					Type: ERR,
					Text: err.Error(),
					Pos:  l.pos,
				}
			}
			l.pos++

			if isSTRING(r) {
				id = append(id, r)
			} else {
				l.in.UnreadRune()
				l.pos--
				return Token{
					Type: STRING,
					Text: string(id),
					Pos:  startpos,
				}
			}
		}
	}
}

func (l *Lexer) skipSpace(initr rune) int {
	posdelta := 0
	if unicode.IsSpace(initr) {
		for {
			r, sz, _ := l.in.ReadRune()
			if sz == 0 {
				return posdelta
			}

			//fmt.Fprintf(os.Stderr, "    next ... %q\n", string(r))
			if !unicode.IsSpace(r) {
				l.in.UnreadRune()
				return posdelta
			}
			posdelta++
		}
	}
	return posdelta
}

var readerPool = sync.Pool{
	New: func() any {
		return bufio.NewReader(nil)
	},
}
