package jsonstream

import (
	"fmt"
	"io"
)

var (
	cURLY_BRACKET_LEFT_BYTES  = []byte("{")
	cURLY_BRACKET_RIGHT_BYTES = []byte("}")
	rECT_BRACKET_LEFT_BYTES   = []byte("[")
	rECT_BRACKET_RIGHT_BYTES  = []byte("]")
	cOLON_BYTES               = []byte(":")
	cOMMA_BYTES               = []byte(",")
	nULL_BYTES                = []byte("null")
	tRUE_BYTES                = []byte("true")
	fALSE_BYTES               = []byte("false")
	qUOTE_BYTES               = []byte("\"")
)

type tokenWriter struct {
	wr io.Writer
}

func NewWriter(wr io.Writer) Writer {
	return Writer(&tokenWriter{wr: wr})
}

func (t *tokenWriter) WriteTokens(tokens ...Token) error {
	for _, token := range tokens {
		err := t.WriteToken(token)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *tokenWriter) WriteToken(token Token) error {

	switch token.Type {
	case TT_OBJECT_START:
		_, err := t.wr.Write(cURLY_BRACKET_LEFT_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_OBJECT_END:
		_, err := t.wr.Write(cURLY_BRACKET_RIGHT_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_ARRAY_START:
		_, err := t.wr.Write(rECT_BRACKET_LEFT_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_ARRAY_END:
		_, err := t.wr.Write(rECT_BRACKET_RIGHT_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_KEY:
		_, err := t.wr.Write(qUOTE_BYTES)
		if err != nil {
			return err
		}

		_, err = t.wr.Write([]byte(token.Value))
		if err != nil {
			return err
		}

		_, err = t.wr.Write(qUOTE_BYTES)
		if err != nil {
			return err
		}

		return nil
	case TT_COLON:
		_, err := t.wr.Write(cOLON_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_COMMA:
		_, err := t.wr.Write(cOMMA_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_STRING_VALUE:
		_, err := t.wr.Write([]byte(fmt.Sprintf("\"%s\"", token.Value)))
		if err != nil {
			return err
		}
		return nil
	case TT_NULL_VALUE:
		_, err := t.wr.Write(nULL_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_TRUE_VALUE:
		_, err := t.wr.Write(tRUE_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_FALSE_VALUE:
		_, err := t.wr.Write(fALSE_BYTES)
		if err != nil {
			return err
		}
		return nil
	case TT_NUMBER_VALUE:
		_, err := t.wr.Write([]byte(token.Value))
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid token type: %d", token.Type)
	}

}

func (t *tokenWriter) Close() error {
	closer, isCloser := t.wr.(io.Closer)
	if isCloser {
		return closer.Close()
	}

	return nil
}
