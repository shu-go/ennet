package ennet_test

import (
	"testing"

	"github.com/shu-go/ennet"
	"github.com/shu-go/gotwant"
)

func TestLexerTx(t *testing.T) {
	t.Run("Straightforward", func(t *testing.T) {
		l := input(`a>b>c`)
		tx := l.Transaction()

		tok := tx.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, `a`)
		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.CHILD)
		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, `b`)

		tx.Rollback()

		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, `a`)
		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.CHILD)
		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, `b`)
	})

	t.Run("Branched1", func(t *testing.T) {
		l := input(`a>b>c`)
		tx := l.Transaction()

		tok := tx.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, `a`)
		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.CHILD)

		txx := tx.Transaction()
		tok = txx.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, `b`)
		tok = txx.Next()
		gotwant.Test(t, tok.Type, ennet.CHILD)

		txx.Rollback()

		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, `b`)
		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.CHILD)

		tx.Rollback()

		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.STRING)
		gotwant.Test(t, tok.Text, `a`)
		tok = tx.Next()
		gotwant.Test(t, tok.Type, ennet.CHILD)
	})
}
