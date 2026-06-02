package ennet

import (
	"fmt"
	"io"
	"sync"
	"unsafe"
)

func unsafeString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

func (l *Lexer) Dump() string {
	return "*lexer dump:\n" +
		fmt.Sprintf("  %+v\n", *l.tokens) +
		fmt.Sprintf("  %+v\n", l.scanpos)
}

type Lexer struct {
	in     []byte
	pos    int
	offset int

	// if len(*tokens) == scanpos: next is read from in
	// else: next is from (*tokens)[scanpos], then scanpos++
	tokens  *[]Token
	scanpos int
}

func NewLexer(b []byte) *Lexer {
	l := lexerPool.Get().(*Lexer)
	ps := tokensPool.Get().(*[]Token)
	*ps = (*ps)[:0]

	l.in = b
	l.pos = 1
	l.offset = 0
	l.tokens = ps
	l.scanpos = 0
	return l
}

func (l *Lexer) Close() {
	tokensPool.Put(l.tokens)
	*l = Lexer{}
	lexerPool.Put(l)
}

func (l *Lexer) Next() Token {
	var tok Token
	scanning := *l.tokens
	if l.scanpos == len(scanning) {
		tok = l.scanNext()
		*l.tokens = append(*l.tokens, tok)
	} else { // l.scanpos < len(scanning)
		tok = scanning[l.scanpos]
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
	scanning := *l.tokens
	if l.scanpos < len(scanning) {
		return scanning[l.scanpos]
	}

	tok := l.scanNext()
	*l.tokens = append(*l.tokens, tok)
	return tok
}

func (l *Lexer) readByte() (byte, error) {
	if l.offset >= len(l.in) {
		return 0, io.EOF
	}
	b := l.in[l.offset]
	l.offset++
	return b, nil
}

func (l *Lexer) unreadByte() {
	if l.offset > 0 {
		l.offset--
	}
}

func (l *Lexer) scanNext() Token {
	startpos := l.pos

	c, err := l.readByte()
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
		startOffset := l.offset
		hasEscape := false
		for {
			c, err = l.readByte()
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
				cc, err := l.readByte()
				if err != nil && err != io.EOF {
					return Token{
						Type: ERR,
						Text: err.Error(),
						Pos:  l.pos,
					}
				}
				if cc == quot {
					l.pos++
					hasEscape = true
				} else if err == io.EOF {
					break
				} else {
					l.unreadByte()
					break
				}
			}
		}

		endOffset := l.offset - 1
		if !hasEscape {
			return Token{
				Type: QTEXT,
				Text: unsafeString(l.in[startOffset:endOffset]),
				Pos:  startpos,
			}
		}

		// Slow path: resolve escapes
		text := make([]byte, 0, endOffset-startOffset)
		l.offset = startOffset
		for l.offset < endOffset {
			b := l.in[l.offset]
			l.offset++
			if b == quot {
				l.offset++
			}
			text = append(text, b)
		}
		l.offset = endOffset + 1
		return Token{
			Type: QTEXT,
			Text: string(text),
			Pos:  startpos,
		}

	case '{':
		startOffset := l.offset
		hasEscape := false
		for {
			c, err = l.readByte()
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
				rr, err := l.readByte()
				if rr == '}' {
					l.pos++
					hasEscape = true
				} else if err == io.EOF {
					break
				} else {
					l.unreadByte()
					break
				}
			}
		}

		endOffset := l.offset - 1
		if !hasEscape {
			return Token{
				Type: TEXT,
				Text: unsafeString(l.in[startOffset:endOffset]),
				Pos:  startpos,
			}
		}

		// Slow path: resolve escapes
		text := make([]byte, 0, endOffset-startOffset)
		l.offset = startOffset
		for l.offset < endOffset {
			b := l.in[l.offset]
			l.offset++
			if b == '}' {
				l.offset++
			}
			text = append(text, b)
		}
		l.offset = endOffset + 1
		return Token{
			Type: TEXT,
			Text: string(text),
			Pos:  startpos,
		}

	default:
		if isSpace(c) {
			l.pos += l.skipSpace(c)
			return l.scanNext()
		}

		// STRING
		startOffset := l.offset - 1
		for {
			c, err = l.readByte()
			if err == io.EOF {
				return Token{
					Type: STRING,
					Text: unsafeString(l.in[startOffset:l.offset]),
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

			if !isSTRING(c) {
				l.unreadByte()
				l.pos--
				return Token{
					Type: STRING,
					Text: unsafeString(l.in[startOffset:l.offset]),
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
			r, err := l.readByte()
			if err == io.EOF {
				return posdelta
			}

			if !isSpace(r) {
				l.unreadByte()
				return posdelta
			}
			posdelta++
		}
	}
	return posdelta
}

var tokensPool = sync.Pool{
	New: func() any {
		s := make([]Token, 0, 16)
		return &s
	},
}

var lexerPool = sync.Pool{
	New: func() any {
		return &Lexer{}
	},
}
