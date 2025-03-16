package ennet

import (
	"errors"
	"strconv"
)

/*
EBNF:

	attr-list = "[", attr, { attr }, "]";
	attr = STRING, ["=", (QTEXT | STRING)];
	id = "#", STRING;
	class = ".", STRING;

	tag-element = STRING, { id | class | attr-list }, [ TEXT ];
	multiplication = "*", NUMBER;

	element = ( tag-element | TEXT ), [multiplication];

	group = "(", abbreviation, ")", [multiplication];

	operator = CHILD | SIBLING | repeatable-operator;
	repeatable-operator = CLIMBUP, {CLIMBUP}

	abbreviation = (group | element), [operator, abbreviation]
*/
type Parser struct {
	lexer   *Lexer
	builder Builder
}

func Parse(b []byte, builder Builder) (parseError error) {
	p := Parser{
		lexer:   NewLexer(b),
		builder: builder,
	}

	defer func() {
		if err, ok := recover().(error); ok {
			parseError = err
		} else if err != nil {
			p.lexer.Close()
			panic(err)
		}
		p.lexer.Close()
	}()

	t := p.lexer.Peek()
	result := false
	switch t.Type {
	case GROUPBEGIN /*tagElement*/, STRING, TEXT:
		result = p.abbreviation()
	default:

	}

	tok := p.lexer.Next()
	if tok.Type == ERR {
		return errors.New("parsing failed because of " + tok.String() + strconv.Itoa(tok.Pos))
	} else if tok.Type != EOF {
		return errors.New("parsing failed because of extra " + tok.String() + strconv.Itoa(tok.Pos))
	}
	if !result {
		p.lexer.Close()
		return errors.New("parse error")
	}

	return nil
}

func (p *Parser) abbreviation() bool {
	t := p.lexer.Peek()
	if t.Type == GROUPBEGIN && p.group() {
		// nop
	} else {
		switch t.Type {
		case /*tagElement*/ STRING, TEXT:
			if !p.element() {
				return false
			}
		default:
			return false
		}
	}

	t = p.lexer.Peek()
	switch t.Type {
	case CHILD, SIBLING /*repeatableOperator*/, CLIMBUP:
		if p.operator() {
			t = p.lexer.Peek()
			switch t.Type {
			case GROUPBEGIN /*tagElement*/, STRING, TEXT:
				return p.abbreviation()
			default:
			}
			return false
		}
	default:
	}

	return true
}

func (p *Parser) element() bool {
	t := p.lexer.Peek()
	if t.Type == STRING && p.tagElement() {
		// nop
	} else {
		tok := p.lexer.Next()
		if tok.Type != TEXT {
			return false
		}

		if err := p.builder.Text(tok.Text); err != nil {
			panic(err)
		}
	}

	t = p.lexer.Peek()
	if t.Type == MULT {
		p.multiplication()
	}

	return true
}

func (p *Parser) tagElement() bool {
	tok := p.lexer.Next()
	if tok.Type != STRING {
		return false
	}

	if err := p.builder.Element(tok.Text); err != nil {
		panic(err)
	}

	for {
		t := p.lexer.Peek()
		switch t.Type {
		case ID:
			if p.id() {
				continue
			}
		case CLASS:
			if p.class() {
				continue
			}
		case ATTRBEGIN:
			if p.attrList() {
				continue
			}
		default:

		}
		break
	}

	tok = p.lexer.Next()
	if tok.Type == TEXT {
		if err := p.builder.Text(tok.Text); err != nil {
			panic(err)
		}

	} else {
		p.lexer.Back()
	}

	return true
}

func (p *Parser) id() bool {
	tok := p.lexer.Next()
	if tok.Type != ID {
		return false
	}

	tok = p.lexer.Next()
	if tok.Type != STRING {
		panic(errors.New("id name is required"))
	}

	if err := p.builder.ID(tok.Text); err != nil {
		panic(err)
	}

	return true
}

func (p *Parser) class() bool {
	tok := p.lexer.Next()
	if tok.Type != CLASS {
		return false
	}

	tok = p.lexer.Next()
	if tok.Type != STRING {
		panic(errors.New("class name is required"))
	}

	if err := p.builder.Class(tok.Text); err != nil {
		panic(err)
	}

	return true
}

func (p *Parser) attr() bool {
	tok := p.lexer.Next()
	if tok.Type != STRING {
		return false
	}

	attrname := tok.Text

	tok = p.lexer.Next()
	if tok.Type != EQ {
		p.lexer.Back()

		if err := p.builder.Attribute(attrname, ""); err != nil {
			panic(err)
		}

		return true
	}

	tok = p.lexer.Next()
	if tok.Type != QTEXT && tok.Type != STRING {
		panic(errors.New("attr value is required"))
		//return false
	}

	if err := p.builder.Attribute(attrname, tok.Text); err != nil {
		panic(err)
	}

	return true
}

func (p *Parser) attrList() bool {
	tok := p.lexer.Next()
	if tok.Type != ATTRBEGIN {
		return false
	}

	t := p.lexer.Peek()
	if t.Type != STRING || !p.attr() {
		panic(errors.New("AttrName as a string is required"))
	}

	for {
		t := p.lexer.Peek()
		if t.Type != STRING || !p.attr() {
			break
		}
	}

	tok = p.lexer.Next()
	if tok.Type != ATTREND {
		panic(errors.New("] is required in the end of attrs"))
		//return false
	}

	return true
}

func (p *Parser) group() bool {
	tok := p.lexer.Next()
	if tok.Type == GROUPBEGIN {
		if err := p.builder.GroupBegin(); err != nil {
			panic(err)
		}

		t := p.lexer.Peek()
		switch t.Type {
		case GROUPBEGIN /*tagElement*/, STRING, TEXT:
			if !p.abbreviation() {
				panic(errors.New("A group or element is required"))
			}
		default:
			panic(errors.New("A group or element is required"))
		}

		tok = p.lexer.Next()
		if tok.Type != GROUPEND {
			panic(errors.New(") is required in the end of a group"))
			//return false
		}
		if err := p.builder.GroupEnd(); err != nil {
			panic(err)
		}
	} else {
		return false
	}

	t := p.lexer.Peek()
	if t.Type == MULT {
		p.multiplication()
	}

	return true
}

func (p *Parser) multiplication() bool {
	tok := p.lexer.Next()
	if tok.Type != MULT {
		return false
	}

	tok = p.lexer.Next()
	if tok.Type != STRING {
		panic(errors.New("a number following * is required"))
		//return false
	}
	count, err := strconv.Atoi(tok.Text)
	if err != nil {
		panic(errors.New("a number following * is required"))
		//return false
	}

	if err := p.builder.Mul(count); err != nil {
		panic(err)
	}

	return true
}

func (p *Parser) operator() bool {
	tok := p.lexer.Next()
	if tok.Type == CHILD {
		if err := p.builder.OpChild(); err != nil {
			panic(err)
		}
		return true

	} else if tok.Type == SIBLING {
		if err := p.builder.OpSibling(); err != nil {
			panic(err)
		}
		return true

	} else {
		p.lexer.Back()
		t := p.lexer.Peek()
		if t.Type == CLIMBUP && p.repeatableOperator() {
			return true
		}
	}
	return false
}

func (p *Parser) repeatableOperator() bool {
	count := 0
	tok := p.lexer.Next()
	if tok.Type == CLIMBUP {
		for {
			if tok.Type == CLIMBUP {
				count++
				tok = p.lexer.Next()
			} else {
				p.lexer.Back()
				break
			}
		}

		if count > 0 {
			if err := p.builder.OpClimbup(count); err != nil {
				panic(err)
			}
		}

		return true
	}

	return false
}
