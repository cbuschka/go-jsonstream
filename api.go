package jsonstream

type Writer interface {
	WriteTokens(tokens ...Token) error
	SetIndent(indent string)
	Close() error
}
