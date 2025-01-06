package ennet

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

func (l *Lexer) Dump() string {
	return "*lexer dump:\n" +
		fmt.Sprintf("  %+v\n", l.scanning) +
		fmt.Sprintf("  %+v\n", l.scanpos)
}

type Lexer struct {
	in  *bytes.Buffer
	pos int

	// if len(scanning) == scanpos: next is read from in
	// else: next is from scannin[scanpos], then scanpos++
	scanning []Token
	scanpos  int
}

func NewLexer(b []byte) *Lexer {
	r := readerPool.Get().(*bytes.Buffer)
	r.Reset()
	r.Write(b)

	ps := tokensPool.Get().(*[]Token)
	//println("new", cap(*ps))
	s := *ps
	s = s[:0]

	return &Lexer{
		in:       r,
		pos:      1,
		scanning: s,
	}
}

func (l *Lexer) Close() {
	s := l.scanning
	ps := &s
	//println("close", cap(*ps))
	tokensPool.Put(ps)

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

func (l *Lexer) Peek() Token {
	if l.scanpos < len(l.scanning) {
		return l.scanning[l.scanpos]
	}

	//if l.scanpos == len(l.scanning) {
	tok := l.scanNext()
	l.scanning = append(l.scanning, tok)
	return tok
	//}

}

func (l *Lexer) scanNext() Token {
	startpos := l.pos

	c, err := l.in.ReadByte()
	if err == io.EOF {
		return Token{
			Type: EOF,
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

	switch c {
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
		quot := c
		text := []byte{}
		for {
			c, err = l.in.ReadByte()
			if err == io.EOF {
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

			if c == quot {
				cc, err := l.in.ReadByte()
				if err != nil && err != io.EOF {
					return Token{
						Type: ERR,
						Text: err.Error(),
						Pos:  l.pos,
					}
				}
				if cc == quot {
					l.pos++
					text = append(text, c)
				} else if err == io.EOF {
					break
				} else {
					l.in.UnreadByte()
					break
				}
			} else {
				text = append(text, c)
			}
		}

		return Token{
			Type: QTEXT,
			Text: string(text),
			Pos:  startpos,
		}

	case '{':
		text := []byte{}
		for {
			c, err = l.in.ReadByte()
			if err == io.EOF {
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

			if c == '}' {
				rr, err := l.in.ReadByte()
				if rr == '}' {
					l.pos++
					text = append(text, c)
				} else if err == io.EOF {
					break
				} else {
					l.in.UnreadByte()
					break
				}
			} else {
				text = append(text, c)
			}
		}

		return Token{
			Type: TEXT,
			Text: string(text),
			Pos:  startpos,
		}

	default:
		if isSpace(c) {
			//fmt.Fprintf(os.Stderr, "skipping ... %q\n", string(r))
			l.pos += l.skipSpace(c)
			return l.scanNext()
		}

		// STRING
		id := []byte{c}
		for {
			c, err = l.in.ReadByte()
			if err == io.EOF {
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

			if isSTRING(c) {
				id = append(id, c)
			} else {
				l.in.UnreadByte()
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

func (l *Lexer) skipSpace(initr byte) int {
	posdelta := 0
	if isSpace(initr) {
		for {
			r, err := l.in.ReadByte()
			if err == io.EOF {
				return posdelta
			}

			//fmt.Fprintf(os.Stderr, "    next ... %q\n", string(r))
			if !isSpace(r) {
				l.in.UnreadByte()
				return posdelta
			}
			posdelta++
		}
	}
	return posdelta
}

var readerPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

var tokensPool = sync.Pool{
	New: func() any {
		//println("new")
		s := make([]Token, 0, 16)
		return &s
	},
}
