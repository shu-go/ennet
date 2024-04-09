package ennet_test

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/shu-go/ennet"
	"github.com/shu-go/gotwant"
)

func input(s string) *ennet.Lexer {
	b := bytes.NewBufferString(s)
	return ennet.NewLexer(b)
}

func test(t *testing.T, l *ennet.Lexer, tokens ...ennet.Token) {
	t.Helper()
	for i, want := range tokens {
		got := l.Next()
		gotwant.Test(t, got.Type, want.Type, gotwant.Desc(strconv.Itoa(i)+" Type"))
		gotwant.Test(t, got.Text, want.Text, gotwant.Desc(strconv.Itoa(i)+" Text"))
	}
}

func TestLexerOne(t *testing.T) {
	t.Run("STRING", func(t *testing.T) {
		l := input("abc")

		tok := l.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, "abc")

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)
	})

	t.Run("TEXT", func(t *testing.T) {
		l := input("{abc}")

		tok := l.Next()
		gotwant.Test(t, tok.Type, ennet.TEXT)
		gotwant.Test(t, tok.Text, "abc")

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)
	})

	t.Run("TEXTEscaped", func(t *testing.T) {
		l := input("{abc{def}}ghi}")

		tok := l.Next()
		gotwant.Test(t, tok.Type, ennet.TEXT)
		gotwant.Test(t, tok.Text, "abc{def}ghi")

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)
	})

	t.Run("QTEXTSingle", func(t *testing.T) {
		l := input("'abc'")

		tok := l.Next()
		gotwant.Test(t, tok.Type, ennet.QTEXT)
		gotwant.Test(t, tok.Text, "abc")

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)
	})

	t.Run("QTEXTDouble", func(t *testing.T) {
		l := input(`"abc"`)

		tok := l.Next()
		gotwant.Test(t, tok.Type, ennet.QTEXT)
		gotwant.Test(t, tok.Text, "abc")

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)
	})

	t.Run("QTEXTEscaped", func(t *testing.T) {
		l := input(`"ab""c"`)

		tok := l.Next()
		gotwant.Test(t, tok.Type, ennet.QTEXT)
		gotwant.Test(t, tok.Text, `ab"c`)

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)

		l = input(`'ab''c'`)

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.QTEXT)
		gotwant.Test(t, tok.Text, `ab'c`)

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)

		l = input(`"ab'c"`)

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.QTEXT)
		gotwant.Test(t, tok.Text, `ab'c`)

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)

		l = input(`'ab"c'`)

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.QTEXT)
		gotwant.Test(t, tok.Text, `ab"c`)

		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)

	})

	t.Run("Others", func(t *testing.T) {
		l := input(">")
		tok := l.Next()
		gotwant.Test(t, tok.Type, ennet.CHILD)

		l = input("+")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.SIBLING)

		l = input("^")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.CLIMBUP)

		l = input("*")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.MULT)

		l = input("(")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.GROUPBEGIN)

		l = input(")")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.GROUPEND)

		l = input("#")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.ID)

		l = input(".")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.CLASS)

		l = input("[")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.ATTRBEGIN)

		l = input("]")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.ATTREND)

		l = input("=")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EQ)

		l = input("")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)

		l = input(" ")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)

		l = input("  ")
		tok = l.Next()
		gotwant.Test(t, tok.Type, ennet.EOF)
	})
}

func TestLexerCombo(t *testing.T) {
	t.Run("CHILDSIBLINGCLIMBUP", func(t *testing.T) {
		l := input("abc>def+ghi^jkl")
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "abc"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "def"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "ghi"},
			ennet.Token{Type: ennet.CLIMBUP},
			ennet.Token{Type: ennet.STRING, Text: "jkl"},
			ennet.Token{Type: ennet.EOF},
		)
	})

	t.Run("GROUPMULT", func(t *testing.T) {
		l := input("abc>(def+ghi)*5")
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "abc"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.GROUPBEGIN},
			ennet.Token{Type: ennet.STRING, Text: "def"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "ghi"},
			ennet.Token{Type: ennet.GROUPEND},
			ennet.Token{Type: ennet.MULT},
			ennet.Token{Type: ennet.STRING, Text: "5"},
			ennet.Token{Type: ennet.EOF},
		)
	})

	t.Run("ATTR", func(t *testing.T) {
		l := input(`a[attr1="abc" attr2="" attr3]`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "a"},
			ennet.Token{Type: ennet.ATTRBEGIN},
			ennet.Token{Type: ennet.STRING, Text: "attr1"},
			ennet.Token{Type: ennet.EQ},
			ennet.Token{Type: ennet.QTEXT, Text: "abc"},
			ennet.Token{Type: ennet.STRING, Text: "attr2"},
			ennet.Token{Type: ennet.EQ},
			ennet.Token{Type: ennet.QTEXT, Text: ""},
			ennet.Token{Type: ennet.STRING, Text: "attr3"},
			ennet.Token{Type: ennet.ATTREND},
			ennet.Token{Type: ennet.EOF},
		)
	})
}

func TestLexerEmmetDocumentation(t *testing.T) {
	t.Run("div>ul>li", func(t *testing.T) {
		l := input(`div>ul>li`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "ul"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "li"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run("div+p+bq", func(t *testing.T) {
		l := input(`div+p+bq`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "p"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "bq"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run("div+div>p>span+em", func(t *testing.T) {
		l := input(`div+div>p>span+em`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "p"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "span"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "em"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run("div+div>p>span+em^bq", func(t *testing.T) {
		l := input(`div+div>p>span+em^bq`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "p"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "span"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "em"},
			ennet.Token{Type: ennet.CLIMBUP},
			ennet.Token{Type: ennet.STRING, Text: "bq"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run("div+div>p>span+em^^^bq", func(t *testing.T) {
		l := input(`div+div>p>span+em^^^bq`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "p"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "span"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "em"},
			ennet.Token{Type: ennet.CLIMBUP},
			ennet.Token{Type: ennet.CLIMBUP},
			ennet.Token{Type: ennet.CLIMBUP},
			ennet.Token{Type: ennet.STRING, Text: "bq"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run("ul>li*5", func(t *testing.T) {
		l := input(`ul>li*5`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "ul"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "li"},
			ennet.Token{Type: ennet.MULT},
			ennet.Token{Type: ennet.STRING, Text: "5"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run("div>(header>ul>li*2>a)+footer>p", func(t *testing.T) {
		l := input(`div>(header>ul>li*2>a)+footer>p`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.GROUPBEGIN},
			ennet.Token{Type: ennet.STRING, Text: "header"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "ul"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "li"},
			ennet.Token{Type: ennet.MULT},
			ennet.Token{Type: ennet.STRING, Text: "2"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "a"},
			ennet.Token{Type: ennet.GROUPEND},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "footer"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "p"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run("(div>dl>(dt+dd)*3)+footer>p", func(t *testing.T) {
		l := input(`(div>dl>(dt+dd)*3)+footer>p`)
		test(t, l,
			ennet.Token{Type: ennet.GROUPBEGIN},
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "dl"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.GROUPBEGIN},
			ennet.Token{Type: ennet.STRING, Text: "dt"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "dd"},
			ennet.Token{Type: ennet.GROUPEND},
			ennet.Token{Type: ennet.MULT},
			ennet.Token{Type: ennet.STRING, Text: "3"},
			ennet.Token{Type: ennet.GROUPEND},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "footer"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "p"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run("div#header+div.page+div#footer.class1.class2.class3", func(t *testing.T) {
		l := input(`div#header+div.page+div#footer.class1.class2.class3`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.ID},
			ennet.Token{Type: ennet.STRING, Text: "header"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.CLASS},
			ennet.Token{Type: ennet.STRING, Text: "page"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "div"},
			ennet.Token{Type: ennet.ID},
			ennet.Token{Type: ennet.STRING, Text: "footer"},
			ennet.Token{Type: ennet.CLASS},
			ennet.Token{Type: ennet.STRING, Text: "class1"},
			ennet.Token{Type: ennet.CLASS},
			ennet.Token{Type: ennet.STRING, Text: "class2"},
			ennet.Token{Type: ennet.CLASS},
			ennet.Token{Type: ennet.STRING, Text: "class3"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run(`td[title="Hello world!" colspan=3]`, func(t *testing.T) {
		l := input(`td[title="Hello world!" colspan=3]`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "td"},
			ennet.Token{Type: ennet.ATTRBEGIN},
			ennet.Token{Type: ennet.STRING, Text: "title"},
			ennet.Token{Type: ennet.EQ},
			ennet.Token{Type: ennet.QTEXT, Text: "Hello world!"},
			ennet.Token{Type: ennet.STRING, Text: "colspan"},
			ennet.Token{Type: ennet.EQ},
			ennet.Token{Type: ennet.STRING, Text: "3"},
			ennet.Token{Type: ennet.ATTREND},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run(`ul>li.item$*5`, func(t *testing.T) {
		l := input(`ul>li.item$*5`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "ul"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "li"},
			ennet.Token{Type: ennet.CLASS},
			ennet.Token{Type: ennet.STRING, Text: "item$"},
			ennet.Token{Type: ennet.MULT},
			ennet.Token{Type: ennet.STRING, Text: "5"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run(`ul>li.item$$$*5`, func(t *testing.T) {
		l := input(`ul>li.item$$$*5`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "ul"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "li"},
			ennet.Token{Type: ennet.CLASS},
			ennet.Token{Type: ennet.STRING, Text: "item$$$"},
			ennet.Token{Type: ennet.MULT},
			ennet.Token{Type: ennet.STRING, Text: "5"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run(`ul>li.item$@-*5`, func(t *testing.T) {
		l := input(`ul>li.item$@-*5`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "ul"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "li"},
			ennet.Token{Type: ennet.CLASS},
			ennet.Token{Type: ennet.STRING, Text: "item$@-"},
			ennet.Token{Type: ennet.MULT},
			ennet.Token{Type: ennet.STRING, Text: "5"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run(`ul>li.item$@-*5`, func(t *testing.T) {
		l := input(`ul>li.item$@-3*5`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "ul"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.STRING, Text: "li"},
			ennet.Token{Type: ennet.CLASS},
			ennet.Token{Type: ennet.STRING, Text: "item$@-3"},
			ennet.Token{Type: ennet.MULT},
			ennet.Token{Type: ennet.STRING, Text: "5"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run(`a{Click me}`, func(t *testing.T) {
		l := input(`a{Click me}`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "a"},
			ennet.Token{Type: ennet.TEXT, Text: "Click me"},
			ennet.Token{Type: ennet.EOF},
		)
	})
	t.Run(`p>{Click }+a{here}+{ to continue}`, func(t *testing.T) {
		l := input(`p>{Click }+a{here}+{ to continue}`)
		test(t, l,
			ennet.Token{Type: ennet.STRING, Text: "p"},
			ennet.Token{Type: ennet.CHILD},
			ennet.Token{Type: ennet.TEXT, Text: "Click "},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.STRING, Text: "a"},
			ennet.Token{Type: ennet.TEXT, Text: "here"},
			ennet.Token{Type: ennet.SIBLING},
			ennet.Token{Type: ennet.TEXT, Text: " to continue"},
			ennet.Token{Type: ennet.EOF},
		)
	})
}

func TestLexerError(t *testing.T) {
	t.Run("EOF", func(t *testing.T) {
		l := input("")
		tok := l.Next()
		gotwant.Test(t, tok, ennet.Token{Type: ennet.EOF})
	})

	t.Run("ERR Suddon EOF", func(t *testing.T) {
		l := input(`{a`)
		tok := l.Next()
		gotwant.Test(t, tok, ennet.Token{Type: ennet.ERR, Text: "sudden EOF"})

		l = input(`"a`)
		tok = l.Next()
		gotwant.Test(t, tok, ennet.Token{Type: ennet.ERR, Text: "sudden EOF"})

		// no ERR
		l = input(`[a`)
		tok = l.Next()
		gotwant.Test(t, tok, ennet.Token{Type: ennet.ATTRBEGIN, Text: ""})
		tok = l.Next()
		gotwant.Test(t, tok, ennet.Token{Type: ennet.STRING, Text: "a"})
		tok = l.Next()
		gotwant.Test(t, tok, ennet.Token{Type: ennet.EOF, Text: ""})
	})
}
