package ennet_test

import (
	"bytes"
	"testing"

	"github.com/shu-go/ennet"
	"github.com/shu-go/gotwant"
)

func TestParser(t *testing.T) {
	t.Run("Element", func(t *testing.T) {
		b := bytes.NewBufferString(`a`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.FirstChild.Data, "a")
	})

	t.Run("Text", func(t *testing.T) {
		b := bytes.NewBufferString(`{hoge}`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  "hoge"`)
	})

	t.Run("Child", func(t *testing.T) {
		b := bytes.NewBufferString(`a>b>c`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.FirstChild.Data, "a")
		gotwant.Test(t, nl.Root.FirstChild.FirstChild.Data, "b")
		gotwant.Test(t, nl.Root.FirstChild.FirstChild.FirstChild.Data, "c")
	})

	t.Run("Sibling", func(t *testing.T) {
		b := bytes.NewBufferString(`a+b+c`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.FirstChild.Data, "a")
		gotwant.Test(t, nl.Root.FirstChild.NextSibling.Data, "b")
		gotwant.Test(t, nl.Root.FirstChild.NextSibling.NextSibling.Data, "c")
	})

	t.Run("IDClassAttrText", func(t *testing.T) {
		b := bytes.NewBufferString(`a#idid.cls.cls2[attr1 attr2="value2" attr3='value3']{text desu}`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.FirstChild.Data, "a")
		gotwant.Test(t, nl.Root.FirstChild.Attribute["id"], "idid")
		gotwant.Test(t, nl.Root.FirstChild.Attribute["class"], "cls cls2")
		gotwant.Test(t, nl.Root.FirstChild.Attribute["attr1"], "")
		gotwant.Test(t, nl.Root.FirstChild.Attribute["attr2"], "value2")
		gotwant.Test(t, nl.Root.FirstChild.Attribute["attr3"], "value3")
		gotwant.Test(t, nl.Root.FirstChild.FirstChild.Data, "text desu")
	})

	t.Run("Family", func(t *testing.T) {
		b := bytes.NewBufferString(`a+b>c^d`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.FirstChild.Data, "a")
		gotwant.Test(t, nl.Root.FirstChild.NextSibling.Data, "b")
		gotwant.Test(t, nl.Root.FirstChild.NextSibling.FirstChild.Data, "c")
		gotwant.Test(t, nl.Root.FirstChild.NextSibling.NextSibling.Data, "d")

		gotwant.Test(t, nl.Root.Dump(), `
  a:element
  b:element
    c:element
  d:element`)
	})
}

func TestParserEmmetDocumentation(t *testing.T) {
	t.Run("div>ul>li", func(t *testing.T) {
		b := bytes.NewBufferString(`div>ul>li`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  div:element
    ul:element
      li:element`)
	})

	t.Run("div+p+bq", func(t *testing.T) {
		b := bytes.NewBufferString(`div+p+bq`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  div:element
  p:element
  bq:element`)
	})

	t.Run("div+div>p>span+em", func(t *testing.T) {
		b := bytes.NewBufferString(`div+div>p>span+em`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  div:element
  div:element
    p:element
      span:element
      em:element`)
	})

	t.Run("div+div>p>span+em^bq", func(t *testing.T) {
		b := bytes.NewBufferString(`div+div>p>span+em^bq`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  div:element
  div:element
    p:element
      span:element
      em:element
    bq:element`)
	})

	t.Run("div+div>p>span+em^^^bq", func(t *testing.T) {
		b := bytes.NewBufferString(`div+div>p>span+em^^^bq`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  div:element
  div:element
    p:element
      span:element
      em:element
  bq:element`)
	})

	t.Run("ul>li*5", func(t *testing.T) {
		b := bytes.NewBufferString(`ul>li*5`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  ul:element
    li:element *5`)
	})

	t.Run("div>(header>ul>li*2>a)+footer>p", func(t *testing.T) {
		b := bytes.NewBufferString(`div>(header>ul>li*2>a)+footer>p`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  div:element
    :group
      header:element
        ul:element
          li:element *2
            a:element
    footer:element
      p:element`)
	})

	t.Run("(div>dl>(dt+dd)*3)+footer>p", func(t *testing.T) {
		b := bytes.NewBufferString(`(div>dl>(dt+dd)*3)+footer>p`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  :group
    div:element
      dl:element
        :group *3
          dt:element
          dd:element
  footer:element
    p:element`)
	})

	t.Run("div#header+div.page+div#footer.class1.class2.class3", func(t *testing.T) {
		b := bytes.NewBufferString(`div#header+div.page+div#footer.class1.class2.class3`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  div:element @id=header
  div:element @class=page
  div:element @class=class1 class2 class3 @id=footer`)
	})

	t.Run(`td[title="Hello world!" colspan=3]`, func(t *testing.T) {
		b := bytes.NewBufferString(`td[title="Hello world!" colspan=3]`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  td:element @colspan=3 @title=Hello world!`)
	})

	t.Run(`ul>li.item$*5`, func(t *testing.T) {
		b := bytes.NewBufferString(`ul>li.item$*5`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  ul:element
    li:element @class=item$ *5`)
	})

	t.Run(`ul>li.item$@-*5`, func(t *testing.T) {
		b := bytes.NewBufferString(`ul>li.item$@-*5`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  ul:element
    li:element @class=item$@- *5`)
	})

	t.Run(`ul>li.item$@3*5`, func(t *testing.T) {
		b := bytes.NewBufferString(`ul>li.item$@3*5`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  ul:element
    li:element @class=item$@3 *5`)
	})

	t.Run(`ul>li.item$@-3*5`, func(t *testing.T) {
		b := bytes.NewBufferString(`ul>li.item$@-3*5`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  ul:element
    li:element @class=item$@-3 *5`)
	})

	t.Run(`a{Click me}`, func(t *testing.T) {
		b := bytes.NewBufferString(`a{Click me}`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  a:element
    "Click me"`)
	})

	t.Run(`p>{Click }+a{here}+{ to continue}`, func(t *testing.T) {
		b := bytes.NewBufferString(`p>{Click }+a{here}+{ to continue}`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  p:element
    "Click "
    a:element
      "here"
    " to continue"`)
	})

	t.Run(`p{Click }+a{here}+{ to continue}`, func(t *testing.T) {
		b := bytes.NewBufferString(`p{Click }+a{here}+{ to continue}`)
		nl := ennet.NewNodeBuilder()

		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		gotwant.Test(t, nl.Root.Dump(), `
  p:element
    "Click "
  a:element
    "here"
  " to continue"`)
	})
}

func TestParserError(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		b := bytes.NewBufferString(`a#`)
		nl := ennet.NewNodeBuilder()
		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "id name is required")

		b = bytes.NewBufferString(`a.`)
		nl = ennet.NewNodeBuilder()
		err = ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "class name is required")

		b = bytes.NewBufferString(`a[attr=]`)
		nl = ennet.NewNodeBuilder()
		err = ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "attr value is required")

		b = bytes.NewBufferString(`a[attr`)
		nl = ennet.NewNodeBuilder()
		err = ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "] is required in the end of attrs")

		// this is OK
		b = bytes.NewBufferString(`a[attr]`)
		nl = ennet.NewNodeBuilder()
		err = ennet.Parse(b, &nl)
		gotwant.TestError(t, err, nil)

		b = bytes.NewBufferString(`a*`)
		nl = ennet.NewNodeBuilder()
		err = ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "a number following * is required")
	})

	t.Run("OperatorFirst", func(t *testing.T) {
		b := bytes.NewBufferString(`+hoge`)
		nl := ennet.NewNodeBuilder()
		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "+")
	})

	t.Run("MulFirst", func(t *testing.T) {
		b := bytes.NewBufferString(`*9`)
		nl := ennet.NewNodeBuilder()
		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "*")
	})

	t.Run("AttrOfText", func(t *testing.T) {
		b := bytes.NewBufferString(`{hoge}[aabc]`)
		nl := ennet.NewNodeBuilder()
		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "[")
	})

	t.Run("WrongOrder", func(t *testing.T) {
		b := bytes.NewBufferString(`a{hoge}[aabc]`)
		nl := ennet.NewNodeBuilder()
		err := ennet.Parse(b, &nl)
		gotwant.TestError(t, err, "[")
	})
}

func init() {
	//rog.DisableDebug()
}
