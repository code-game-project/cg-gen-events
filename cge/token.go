package cge

type TokenType string

const (
	EVENT TokenType = "EVENT"
	TYPE  TokenType = "TYPE"

	STRING  TokenType = "STRING"
	BOOL    TokenType = "BOOL"
	INT32   TokenType = "INT32"
	INT64   TokenType = "INT64"
	BIGINT  TokenType = "BIGINT"
	FLOAT32 TokenType = "FLOAT32"
	FLOAT64 TokenType = "FLOAT64"

	MAP  TokenType = "MAP"
	LIST TokenType = "LIST"

	IDENTIFIER TokenType = "IDENTIFIER"

	OPEN_CURLY  TokenType = "OPEN_CURLY"
	CLOSE_CURLY TokenType = "CLOSE_CURLY"
	COLON       TokenType = "COLON"
	COMMA       TokenType = "COMMA"
	GREATER     TokenType = "GREATER"
	LESS        TokenType = "LESS"

	COMMENT TokenType = "COMMENT"

	EOF TokenType = "EOF"
)

type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
	Column int
}
