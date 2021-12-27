package jsonstream

type Writer interface {
	WriteTokens(tokens ...Token) error
	Close() error
}
