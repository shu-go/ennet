package ennet

import (
	"errors"
	"io"
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

func Parse(in io.Reader, builder Builder) (parseError error) {
	defer func() {
		if err, ok := recover().(error); ok {
			parseError = err
		} else if err != nil {
			panic(err)
		}
	}()

	p := Parser{
		lexer:   NewLexer(in),
		builder: builder,
	}
	result := p.precheck(GROUPBEGIN /*tagElement*/, STRING, TEXT) && p.abbreviation()

	tok := p.lexer.Next()
	if tok.Type == ERR {
		return errors.New("parsing failed because of " + tok.String() + strconv.Itoa(tok.Pos))
	} else if tok.Type != EOF {
		return errors.New("parsing failed because of extra " + tok.String() + strconv.Itoa(tok.Pos))
	}
	if !result {
		return errors.New("parse error")
	}

	p.lexer.Close()

	return nil
}

func (p *Parser) precheck(t ...TokenType) bool {
	tok := p.lexer.Peek()

	for i := 0; i < len(t); i++ {
		if tok.Type == t[i] {
			return true
		}
	}
	return false
}

func (p *Parser) abbreviation() bool {
	if p.precheck(GROUPBEGIN) && p.group() {
		// nop
	} else {
		if p.precheck( /*tagElement*/ STRING, TEXT) && p.element() {
			// nop
		} else {
			return false
		}
	}

	if p.precheck(CHILD, SIBLING /*repeatableOperator*/, CLIMBUP) && p.operator() {
		return p.precheck(GROUPBEGIN /*tagElement*/, STRING, TEXT) && p.abbreviation()
	}

	return true
}

func (p *Parser) element() bool {
	if p.precheck(STRING) && p.tagElement() {
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

	if p.precheck(MULT) {
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
		if p.precheck(ID) && p.id() {
			continue
		}
		if p.precheck(CLASS) && p.class() {
			continue
		}
		if p.precheck(ATTRBEGIN) && p.attrList() {
			continue
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

	if !p.precheck(STRING) || !p.attr() {
		panic(errors.New("AttrName as a string is required"))
	}

	for {
		if !p.precheck(STRING) || !p.attr() {
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

		if !p.precheck(GROUPBEGIN /*tagElement*/, STRING, TEXT) || !p.abbreviation() {
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

	if p.precheck(MULT) {
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
		if p.precheck(CLIMBUP) && p.repeatableOperator() {
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
