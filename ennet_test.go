package ennet_test

import (
	"fmt"
	"testing"

	"github.com/shu-go/ennet"
	"github.com/shu-go/gotwant"
)

func Example() {
	expanded, _ := ennet.Expand("ul>li.item-${ITEM$}*3")
	fmt.Println(expanded)

	//Output:
	// <ul><li class="item-1">ITEM1</li><li class="item-2">ITEM2</li><li class="item-3">ITEM3</li></ul>
}

func TestEnnet(t *testing.T) {
	t.Run("Text", func(t *testing.T) {
		s, err := ennet.Expand(`{hoge}`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `hoge`)
	})

	t.Run("Texts", func(t *testing.T) {
		s, err := ennet.Expand(`{hoge}+{fuga piyo}`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `hogefuga piyo`)
	})

	t.Run("Element", func(t *testing.T) {
		s, err := ennet.Expand(`a`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a />`)

		s, err = ennet.Expand(`a#id1.classA[attr1 attr2=2]`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a attr1="" attr2="2" class="classA" id="id1" />`)

		s, err = ennet.Expand(`a#id1.classA[attr1 attr2=2]{text desu}`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a attr1="" attr2="2" class="classA" id="id1">text desu</a>`)
	})

	t.Run("Elements", func(t *testing.T) {
		s, err := ennet.Expand(`a+b`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a /><b />`)

		s, err = ennet.Expand(`a>b`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a><b /></a>`)

		s, err = ennet.Expand(`a>b+c`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a><b /><c /></a>`)

		s, err = ennet.Expand(`a>b+c^d{d text}`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a><b /><c /></a><d>d text</d>`)
	})

	t.Run("Group", func(t *testing.T) {
		s, err := ennet.Expand(`(a+b)`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a /><b />`)

		s, err = ennet.Expand(`a>(b>c>d>e)+f`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a><b><c><d><e /></d></c></b><f /></a>`)
	})

	t.Run("Mul", func(t *testing.T) {
		s, err := ennet.Expand(`(a+b)*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a /><b /><a /><b /><a /><b /><a /><b /><a /><b />`)

		s, err = ennet.Expand(`(a{$}+b{$$@-})*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a>1</a><b>05</b><a>2</a><b>04</b><a>3</a><b>03</b><a>4</a><b>02</b><a>5</a><b>01</b>`)

		s, err = ennet.Expand(`a{0+$=$$}*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a>0+1=01</a><a>0+2=02</a><a>0+3=03</a><a>0+4=04</a><a>0+5=05</a>`)

		s, err = ennet.Expand(`a[attr1-$="$$$@101"]*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a attr1-1="101" /><a attr1-2="102" /><a attr1-3="103" /><a attr1-4="104" /><a attr1-5="105" />`)
	})
}

func TestEnnetDocumentation(t *testing.T) {
	t.Run(`div>ul>li`, func(t *testing.T) {
		s, err := ennet.Expand(`div>ul>li`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<div><ul><li /></ul></div>`)
	})

	t.Run(`div+p+bq`, func(t *testing.T) {
		s, err := ennet.Expand(`div+p+bq`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<div /><p /><bq />`)
	})

	t.Run(`div+div>p>span+em `, func(t *testing.T) {
		s, err := ennet.Expand(`div+div>p>span+em`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<div /><div><p><span /><em /></p></div>`)
	})

	t.Run(`div+div>p>span+em^bq`, func(t *testing.T) {
		s, err := ennet.Expand(`div+div>p>span+em^bq`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<div /><div><p><span /><em /></p><bq /></div>`)
	})

	t.Run(`div+div>p>span+em^^^bq`, func(t *testing.T) {
		s, err := ennet.Expand(`div+div>p>span+em^^^bq`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<div /><div><p><span /><em /></p></div><bq />`)
	})

	t.Run(`ul>li*5`, func(t *testing.T) {
		s, err := ennet.Expand(`ul>li*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<ul><li /><li /><li /><li /><li /></ul>`)
	})

	t.Run(`div>(header>ul>li*2>a)+footer>p`, func(t *testing.T) {
		s, err := ennet.Expand(`div>(header>ul>li*2>a)+footer>p`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<div><header><ul><li><a /></li><li><a /></li></ul></header><footer><p /></footer></div>`)
	})

	t.Run(`(div>dl>(dt+dd)*3)+footer>p`, func(t *testing.T) {
		s, err := ennet.Expand(`(div>dl>(dt+dd)*3)+footer>p`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<div><dl><dt /><dd /><dt /><dd /><dt /><dd /></dl></div><footer><p /></footer>`)
	})

	t.Run(`div#header+div.page+div#footer.class1.class2.class3`, func(t *testing.T) {
		s, err := ennet.Expand(`div#header+div.page+div#footer.class1.class2.class3`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<div id="header" /><div class="page" /><div class="class1 class2 class3" id="footer" />`)
	})

	t.Run(`td[title="Hello world!" colspan=3]`, func(t *testing.T) {
		s, err := ennet.Expand(`td[title="Hello world!" colspan=3]`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<td colspan="3" title="Hello world!" />`)
	})

	t.Run(`ul>li.item$*5`, func(t *testing.T) {
		s, err := ennet.Expand(`ul>li.item$*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<ul><li class="item1" /><li class="item2" /><li class="item3" /><li class="item4" /><li class="item5" /></ul>`)
	})

	t.Run(`ul>li.item$$$*5`, func(t *testing.T) {
		s, err := ennet.Expand(`ul>li.item$$$*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<ul><li class="item001" /><li class="item002" /><li class="item003" /><li class="item004" /><li class="item005" /></ul>`)
	})

	t.Run(`ul>li.item$@-*5`, func(t *testing.T) {
		s, err := ennet.Expand(`ul>li.item$@-*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<ul><li class="item5" /><li class="item4" /><li class="item3" /><li class="item2" /><li class="item1" /></ul>`)
	})

	t.Run(`ul>li.item$@3*5`, func(t *testing.T) {
		s, err := ennet.Expand(`ul>li.item$@3*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<ul><li class="item3" /><li class="item4" /><li class="item5" /><li class="item6" /><li class="item7" /></ul>`)
	})

	t.Run(`ul>li.item$@-3*5`, func(t *testing.T) {
		s, err := ennet.Expand(`ul>li.item$@-3*5`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<ul><li class="item7" /><li class="item6" /><li class="item5" /><li class="item4" /><li class="item3" /></ul>`)
	})

	t.Run(`a{Click me}`, func(t *testing.T) {
		s, err := ennet.Expand(`a{Click me}`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<a>Click me</a>`)
	})

	t.Run(`p>{Click }+a{here}+{ to continue}`, func(t *testing.T) {
		s, err := ennet.Expand(`p>{Click }+a{here}+{ to continue}`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<p>Click <a>here</a> to continue</p>`)
	})

	t.Run(`p{Click }+a{here}+{ to continue}`, func(t *testing.T) {
		s, err := ennet.Expand(`p{Click }+a{here}+{ to continue}`)
		gotwant.TestError(t, err, nil)
		gotwant.Test(t, s, `<p>Click </p><a>here</a> to continue`)
	})
}

func BenchmarkEnnet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ennet.Expand(`div>(header>ul>li*2>a)+footer>p`)
	}
}
