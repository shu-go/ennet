package ennet

// LexerTx can not be used in pallarel
type LexerTx struct {
	lexer *Lexer
	count int
}

func (tx *LexerTx) Transaction() LexerTx {
	return LexerTx{
		lexer: tx.lexer,
		count: 0,
	}
}

func (tx *LexerTx) Next() Token {
	tx.count++
	return tx.lexer.Next()
}

func (tx *LexerTx) Back() {
	if tx.count > 0 {
		tx.count--
	}
	tx.lexer.Back()
}

func (tx *LexerTx) Rollback() {
	for tx.count > 0 {
		tx.Back()
	}
}
