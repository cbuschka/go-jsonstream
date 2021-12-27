package jsonstream

type TokenType int

const (
	TT_OBJECT_START TokenType = iota
	TT_OBJECT_END
	TT_ARRAY_START
	TT_ARRAY_END
	TT_KEY
	TT_COLON
	TT_COMMA
	TT_STRING_VALUE
	TT_NULL_VALUE
	TT_TRUE_VALUE
	TT_FALSE_VALUE
	TT_NUMBER_VALUE
)

type Token struct {
	Type  TokenType
	Value string
}