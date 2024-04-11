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
	lexer    *Lexer
	listener Listener
}

func Parse(in io.Reader, listener Listener) (parseError error) {
	defer func() {
		if err, ok := recover().(error); ok {
			parseError = err
		} else if err != nil {
			panic(err)
		}
	}()

	p := Parser{
		lexer:    NewLexer(in),
		listener: listener,
	}
	tx := p.lexer.Transaction()
	result := p.precheck(GROUPBEGIN /*tagElement*/, STRING, TEXT) && p.abbreviation(&tx)

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
	tok := p.lexer.Next()
	p.lexer.Back()

	for _, tt := range t {
		if tok.Type == tt {
			return true
		}
	}
	return false
}

func (p *Parser) abbreviation(tx *LexerTx) bool {
	debug("abbreviation", "TRY", "group")
	if p.precheck(GROUPBEGIN) && p.group(tx) {
		debug("abbreviation", "group")

	} else {
		debug("abbreviation", "!group")

		debug("abbreviation", "TRY", "element")
		if p.precheck( /*tagElement*/ STRING, TEXT) && p.element(tx) {
			debug("abbreviation", "element")
		} else {
			debug("abbreviation", "!element")
			return false
		}
	}

	debug("abbreviation", "TRY", "operator")
	if p.precheck(CHILD, SIBLING /*repeatableOperator*/, CLIMBUP) && p.operator(tx) {
		debug("abbreviation", "operator")

		debug("abbreviation", "TRY", "abbreviation")
		return p.precheck(GROUPBEGIN /*tagElement*/, STRING, TEXT) && p.abbreviation(tx)
	}

	return true
}

func (p *Parser) element(tx *LexerTx) bool {
	debug("element", "TRY", "tagElement")
	if p.precheck(STRING) && p.tagElement(tx) {
		debug("element", "tagElement")
	} else {
		debug("element", "TRY", "TEXT")
		tok := tx.Next()
		if tok.Type != TEXT {
			return false
		}
		debug("element", "TEXT", tok.Text)

		if err := p.listener.Text(tok.Text); err != nil {
			panic(err)
		}
	}

	debug("element", "TRY(maybe)", "multiplication")
	if p.precheck(MULT) {
		p.multiplication(tx)
	}

	return true
}

func (p *Parser) tagElement(tx *LexerTx) bool {
	debug("tagElement", "TRY", "STRING")
	tok := tx.Next()
	if tok.Type != STRING {
		debug("tagElement", "!STRING", tok.Type)
		return false
	}
	debug("tagElement", "STRING", tok.String())

	if err := p.listener.Element(tok.Text); err != nil {
		panic(err)
	}

	for {
		debug("tagElement", "TRY", "id")
		//debug(p.lexer.Dump())
		if p.precheck(ID) && p.id(tx) {
			debug("tagElement", "id")
			continue
		}
		debug("tagElement", "TRY", "class")
		//debug(p.lexer.Dump())
		if p.precheck(CLASS) && p.class(tx) {
			debug("tagElement", "class")
			continue
		}
		debug("tagElement", "TRY", "attrList")
		//debug(p.lexer.Dump())
		if p.precheck(ATTRBEGIN) && p.attrList(tx) {
			debug("tagElement", "attrList")
			continue
		}
		//debug(p.lexer.Dump())
		break
	}

	debug("tagElement", "TRY", "TEXT")
	tok = tx.Next()
	if tok.Type == TEXT {
		debug("tagElement", "TEXT", tok.String())

		if err := p.listener.Text(tok.Text); err != nil {
			panic(err)
		}

	} else {
		tx.Back()
	}

	return true
}

func (p *Parser) id(tx *LexerTx) bool {
	debug("id", "TRY", "ID")
	tok := tx.Next()
	if tok.Type != ID {
		debug("id", "!ID", tok.Type)
		return false
	}

	debug("id", "TRY", "STRING")
	tok = tx.Next()
	if tok.Type != STRING {
		debug("id", "!STRING")
		panic(errors.New("id name is required"))
		//return false
	}

	if err := p.listener.ID(tok.Text); err != nil {
		panic(err)
	}

	return true
}

func (p *Parser) class(tx *LexerTx) bool {
	debug("class", "TRY", "CLASS")
	tok := tx.Next()
	if tok.Type != CLASS {
		debug("class", "!CLASS", tok.String())
		return false
	}

	debug("class", "TRY", "STRING")
	tok = tx.Next()
	if tok.Type != STRING {
		debug("class", "!STRING")
		panic(errors.New("class name is required"))
		//return false
	}

	if err := p.listener.Class(tok.Text); err != nil {
		panic(err)
	}

	return true
}

func (p *Parser) attr(tx *LexerTx) bool {
	debug("attr", "TRY", "STRING")
	tok := tx.Next()
	if tok.Type != STRING {
		return false
	}

	attrname := tok.Text

	debug("attr", "TRY", "EQ")
	tok = tx.Next()
	if tok.Type != EQ {
		tx.Back()

		if err := p.listener.Attribute(attrname, ""); err != nil {
			panic(err)
		}

		return true
	}

	debug("attr", "TRY", "QTEXT")
	tok = tx.Next()
	if tok.Type != QTEXT && tok.Type != STRING {
		panic(errors.New("attr value is required"))
		//return false
	}

	if err := p.listener.Attribute(attrname, tok.Text); err != nil {
		panic(err)
	}

	return true
}

func (p *Parser) attrList(tx *LexerTx) bool {
	debug("attrList", "TRY", "ATTRBEGIN")
	tok := tx.Next()
	if tok.Type != ATTRBEGIN {
		debug("attrList", "!ATTRBEGIN", tok.String())
		return false
	}

	if !p.precheck(STRING) || !p.attr(tx) {
		panic(errors.New("AttrName as a string is required"))
	}

	for {
		if !p.precheck(STRING) || !p.attr(tx) {
			break
		}
	}

	tok = tx.Next()
	if tok.Type != ATTREND {
		panic(errors.New("] is required in the end of attrs"))
		//return false
	}

	return true
}

func (p *Parser) group(tx *LexerTx) bool {
	tok := tx.Next()
	if tok.Type == GROUPBEGIN {
		if err := p.listener.GroupBegin(); err != nil {
			panic(err)
		}

		if !p.precheck(GROUPBEGIN /*tagElement*/, STRING, TEXT) || !p.abbreviation(tx) {
			panic(errors.New("A group or element is required"))
		}

		tok = tx.Next()
		if tok.Type != GROUPEND {
			panic(errors.New(") is required in the end of a group"))
			//return false
		}
		if err := p.listener.GroupEnd(); err != nil {
			panic(err)
		}
	} else {
		return false
	}

	if p.precheck(MULT) {
		p.multiplication(tx)
	}

	return true
}

func (p *Parser) multiplication(tx *LexerTx) bool {
	tok := tx.Next()
	if tok.Type != MULT {
		return false
	}

	tok = tx.Next()
	if tok.Type != STRING {
		panic(errors.New("a number following * is required"))
		//return false
	}
	count, err := strconv.Atoi(tok.Text)
	if err != nil {
		panic(errors.New("a number following * is required"))
		//return false
	}

	if err := p.listener.Mul(count); err != nil {
		panic(err)
	}

	return true
}

func (p *Parser) operator(tx *LexerTx) bool {
	tok := tx.Next()
	if tok.Type == CHILD {
		if err := p.listener.OpChild(); err != nil {
			panic(err)
		}
		return true

	} else if tok.Type == SIBLING {
		if err := p.listener.OpSibling(); err != nil {
			panic(err)
		}
		return true

	} else {
		tx.Back()
		if p.precheck(CLIMBUP) && p.repeatableOperator(tx) {
			return true
		}
	}
	return false
}

func (p *Parser) repeatableOperator(tx *LexerTx) bool {
	count := 0
	debug("rOperator", "TRY", "CLIMBUP")
	tok := tx.Next()
	if tok.Type == CLIMBUP {
		debug("rOperator", "CLIMBUP")
		for {
			if tok.Type == CLIMBUP {
				count++
				tok = tx.Next()
			} else {
				tx.Back()
				break
			}
		}

		debug("rOperator", "CLIMBUP", count)
		if count > 0 {
			if err := p.listener.OpClimbup(count); err != nil {
				panic(err)
			}
		}

		return true
	}

	return false
}
