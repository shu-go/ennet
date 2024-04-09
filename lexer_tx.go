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
	tx.count++
	if tx.parent != nil {
		tx.parent.Inc()
	}
}

func (tx *LexerTx) Dec() {
	if tx.count <= 0 {
		return
	}
	tx.count--
	if tx.parent != nil {
		tx.parent.Dec()
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
