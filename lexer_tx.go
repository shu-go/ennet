package ennet

// LexerTx can not be used in pallarel
type LexerTx struct {
	lexer *Lexer

	parent *LexerTx
	count  int
}

func (tx *LexerTx) Transaction() LexerTx {
	return LexerTx{
		lexer:  tx.lexer,
		parent: tx,
		count:  0,
	}
}

func (tx *LexerTx) Inc() {
	curr := tx
	for curr != nil {
		tx.count++
		curr = curr.parent
	}
}

func (tx *LexerTx) Dec() {
	curr := tx
	for curr != nil {
		if tx.count > 0 {
			tx.count--
		}
		curr = curr.parent
	}
}

func (tx *LexerTx) Next() Token {
	tx.Inc()
	return tx.lexer.Next()
}

func (tx *LexerTx) Back() {
	tx.Dec()
	tx.lexer.Back()
}

func (tx *LexerTx) Rollback() {
	for tx.count > 0 {
		tx.Back()
	}
}
