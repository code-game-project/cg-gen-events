package cge

type TokenType string

const (
	NAME    TokenType = "NAME"
	VERSION TokenType = "VERSION"

	CONFIG  TokenType = "CONFIG"
	COMMAND TokenType = "COMMAND"
	EVENT   TokenType = "EVENT"
	TYPE    TokenType = "TYPE"
	ENUM    TokenType = "ENUM"

	STRING  TokenType = "STRING"
	BOOL    TokenType = "BOOL"
	INT32   TokenType = "INT32"
	INT64   TokenType = "INT64"
	BIGINT  TokenType = "BIGINT"
	FLOAT32 TokenType = "FLOAT32"
	FLOAT64 TokenType = "FLOAT64"

	MAP  TokenType = "MAP"
	LIST TokenType = "LIST"

	IDENTIFIER     TokenType = "IDENTIFIER"
	VERSION_NUMBER TokenType = "VERSION_NUMBER"

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
