package internal

import (
	"fmt"
	"io"
	"strconv"
)

type tokenWriterState int

type tokenWriterStateStack []tokenWriterState

func (s *tokenWriterStateStack) Peek() tokenWriterState {
	if len(*s) == 0 {
		panic(fmt.Errorf("empty stack"))
	}

	return (*s)[len(*s)-1]
}

func (s *tokenWriterStateStack) IsEmpty() bool {
	return len(*s) == 0
}

func (s *tokenWriterStateStack) Pop() tokenWriterState {
	state := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return state
}

func (s *tokenWriterStateStack) Push(state tokenWriterState) {
	*s = append(*s, state)
}

func (s *tokenWriterStateStack) Replace(state tokenWriterState) {
	(*s)[len(*s)-1] = state
}

const (
	TWS_INITIAL tokenWriterState = iota
	TWS_IN_OBJECT
	TWS_IN_OBJECT_KEY_SEEN
	TWS_IN_OBJECT_COLON_SEEN
	TWS_IN_OBJECT_PAIR_SEEN
	TWS_IN_OBJECT_COMMA_SEEN
	TWS_IN_ARRAY
	TWS_IN_ARRAY_ITEM_SEEN
	TWS_IN_ARRAY_COMMA_SEEN
	TWS_END
)

var tokenWriterStateNames = []string{"TWS_INITIAL", "TWS_IN_OBJECT", "TWS_IN_OBJECT_KEY_SEEN", "TWS_IN_OBJECT_COLON_SEEN", "TWS_IN_OBJECT_PAIR_SEEN", "TWS_IN_OBJECT_COMMA_SEEN", "TWS_IN_ARRAY", "TWS_IN_ARRAY_ITEM_SEEN", "TWS_IN_ARRAY_COMMA_SEEN", "TWS_END"}

func (s tokenWriterState) Name() string {
	return tokenWriterStateNames[s]
}

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
	lINE_BREAK_BYTES          = []byte("\n")
	sPACE_BYTES               = []byte(" ")
)

type TokenWriter struct {
	wr          io.Writer
	indent      string
	indentLevel int
	stateStack  tokenWriterStateStack
}

func NewTokenWriter(wr io.Writer) *TokenWriter {
	return &TokenWriter{wr: wr, indent: "", indentLevel: 0, stateStack: tokenWriterStateStack{TWS_INITIAL}}
}

func (t *TokenWriter) SetIndent(indent string) {
	t.indent = indent
}

func (t *TokenWriter) WriteTokens(tokens ...Token) error {
	for _, token := range tokens {
		err := t.WriteToken(token)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TokenWriter) checkTokenAllowed(currTokenType TokenType, allowedStates ...tokenWriterState) error {
	currState := t.stateStack.Peek()
	for _, allowedState := range allowedStates {
		if allowedState == currState {
			return nil
		}
	}

	return fmt.Errorf("%s not allowed in %s", currTokenType.Name(), currState.Name())
}

func (t *TokenWriter) WriteToken(token Token) error {

	err := t.addMissingTokens(token)
	if err != nil {
		return err
	}

	switch token.Type {
	case TT_OBJECT_START:
		err := t.checkTokenAllowed(token.Type, TWS_INITIAL, TWS_IN_OBJECT_COLON_SEEN, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN)
		if err != nil {
			return err
		}

		_, err = t.wr.Write(cURLY_BRACKET_LEFT_BYTES)
		if err != nil {
			return err
		}
		if t.indent != "" {
			t.indentLevel++
		}
		t.stateStack.Push(TWS_IN_OBJECT)
		return nil
	case TT_OBJECT_END:
		err := t.checkTokenAllowed(token.Type, TWS_IN_OBJECT, TWS_IN_OBJECT_PAIR_SEEN)
		if err != nil {
			return err
		}

		if t.indent != "" {
			_, err := t.wr.Write(lINE_BREAK_BYTES)
			if err != nil {
				return err
			}
			t.indentLevel--

			err = t.writeIndent()
			if err != nil {
				return err
			}
		}

		_, err = t.wr.Write(cURLY_BRACKET_RIGHT_BYTES)
		if err != nil {
			return err
		}
		_ = t.stateStack.Pop()
		if t.stateStack.Peek() == TWS_INITIAL {
			t.stateStack.Replace(TWS_END)
		} else if t.stateStack.Peek() == TWS_IN_OBJECT_COLON_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_PAIR_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		}
		return nil
	case TT_ARRAY_START:
		err := t.checkTokenAllowed(token.Type, TWS_INITIAL, TWS_IN_OBJECT_COLON_SEEN, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN)
		if err != nil {
			return err
		}

		_, err = t.wr.Write(rECT_BRACKET_LEFT_BYTES)
		if err != nil {
			return err
		}
		if t.indent != "" {
			t.indentLevel++
		}
		t.stateStack.Push(TWS_IN_ARRAY)
		return nil
	case TT_ARRAY_END:
		err := t.checkTokenAllowed(token.Type, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN, TWS_IN_ARRAY_ITEM_SEEN)
		if err != nil {
			return err
		}

		if t.indent != "" {
			_, err := t.wr.Write(lINE_BREAK_BYTES)
			if err != nil {
				return err
			}
			t.indentLevel--

			err = t.writeIndent()
			if err != nil {
				return err
			}
		}

		_, err = t.wr.Write(rECT_BRACKET_RIGHT_BYTES)
		if err != nil {
			return err
		}

		_ = t.stateStack.Pop()
		if t.stateStack.Peek() == TWS_INITIAL {
			t.stateStack.Replace(TWS_END)
		} else if t.stateStack.Peek() == TWS_IN_OBJECT_COLON_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_PAIR_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		}

		return nil
	case TT_KEY:
		err := t.checkTokenAllowed(token.Type, TWS_IN_OBJECT, TWS_IN_OBJECT_COMMA_SEEN)
		if err != nil {
			return err
		}

		if t.indent != "" {
			_, err := t.wr.Write(lINE_BREAK_BYTES)
			if err != nil {
				return err
			}

			err = t.writeIndent()
			if err != nil {
				return err
			}
		}

		_, err = t.wr.Write(qUOTE_BYTES)
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
		t.stateStack.Replace(TWS_IN_OBJECT_KEY_SEEN)
		return nil
	case TT_COLON:
		err := t.checkTokenAllowed(token.Type, TWS_IN_OBJECT_KEY_SEEN)
		if err != nil {
			return err
		}
		_, err = t.wr.Write(cOLON_BYTES)
		if err != nil {
			return err
		}

		if t.indent != "" {
			_, err = t.wr.Write(sPACE_BYTES)
			if err != nil {
				return err
			}
		}

		t.stateStack.Replace(TWS_IN_OBJECT_COLON_SEEN)
		return nil
	case TT_COMMA:
		err := t.checkTokenAllowed(token.Type, TWS_IN_OBJECT_PAIR_SEEN, TWS_IN_ARRAY_ITEM_SEEN)
		if err != nil {
			return err
		}

		_, err = t.wr.Write(cOMMA_BYTES)
		if err != nil {
			return err
		}

		if t.stateStack.Peek() == TWS_IN_OBJECT_PAIR_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_COMMA_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_ITEM_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_COMMA_SEEN)
		}

		return nil
	case TT_STRING_VALUE:
		err := t.checkTokenAllowed(token.Type, TWS_INITIAL, TWS_IN_OBJECT_COLON_SEEN, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN)
		if err != nil {
			return err
		}

		if t.stateStack.Peek() == TWS_IN_ARRAY || t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			if t.indent != "" {
				_, err := t.wr.Write(lINE_BREAK_BYTES)
				if err != nil {
					return err
				}

				err = t.writeIndent()
				if err != nil {
					return err
				}
			}
		}

		_, err = t.wr.Write([]byte(fmt.Sprintf("\"%s\"", token.Value)))
		if err != nil {
			return err
		}

		if t.stateStack.Peek() == TWS_INITIAL {
			t.stateStack.Replace(TWS_END)
		} else if t.stateStack.Peek() == TWS_IN_OBJECT_COLON_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_PAIR_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		}

		return nil
	case TT_NULL_VALUE:
		err := t.checkTokenAllowed(token.Type, TWS_INITIAL, TWS_IN_OBJECT_COLON_SEEN, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN)
		if err != nil {
			return err
		}

		_, err = t.wr.Write(nULL_BYTES)
		if err != nil {
			return err
		}

		if t.stateStack.Peek() == TWS_INITIAL {
			t.stateStack.Replace(TWS_END)
		} else if t.stateStack.Peek() == TWS_IN_OBJECT_COLON_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_PAIR_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		}

		return nil
	case TT_TRUE_VALUE:
		err := t.checkTokenAllowed(token.Type, TWS_INITIAL, TWS_IN_OBJECT_COLON_SEEN, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN)
		if err != nil {
			return err
		}

		_, err = t.wr.Write(tRUE_BYTES)
		if err != nil {
			return err
		}

		if t.stateStack.Peek() == TWS_INITIAL {
			t.stateStack.Replace(TWS_END)
		} else if t.stateStack.Peek() == TWS_IN_OBJECT_COLON_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_PAIR_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		}

		return nil
	case TT_FALSE_VALUE:
		err := t.checkTokenAllowed(token.Type, TWS_INITIAL, TWS_IN_OBJECT_COLON_SEEN, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN)
		if err != nil {
			return err
		}

		_, err = t.wr.Write(fALSE_BYTES)
		if err != nil {
			return err
		}

		if t.stateStack.Peek() == TWS_INITIAL {
			t.stateStack.Replace(TWS_END)
		} else if t.stateStack.Peek() == TWS_IN_OBJECT_COLON_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_PAIR_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		}

		return nil
	case TT_NUMBER_VALUE:
		err := t.checkTokenAllowed(token.Type, TWS_INITIAL, TWS_IN_OBJECT_COLON_SEEN, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN)
		if err != nil {
			return err
		}

		_, err = t.wr.Write([]byte(token.Value))
		if err != nil {
			return err
		}

		if t.stateStack.Peek() == TWS_INITIAL {
			t.stateStack.Replace(TWS_END)
		} else if t.stateStack.Peek() == TWS_IN_OBJECT_COLON_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_PAIR_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		}

		return nil
	case TT_INTEGER_VALUE:
		err := t.checkTokenAllowed(token.Type, TWS_INITIAL, TWS_IN_OBJECT_COLON_SEEN, TWS_IN_ARRAY, TWS_IN_ARRAY_COMMA_SEEN)
		if err != nil {
			return err
		}

		_, err = t.wr.Write([]byte(token.Value))
		if err != nil {
			return err
		}

		if t.stateStack.Peek() == TWS_INITIAL {
			t.stateStack.Replace(TWS_END)
		} else if t.stateStack.Peek() == TWS_IN_OBJECT_COLON_SEEN {
			t.stateStack.Replace(TWS_IN_OBJECT_PAIR_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		} else if t.stateStack.Peek() == TWS_IN_ARRAY_COMMA_SEEN {
			t.stateStack.Replace(TWS_IN_ARRAY_ITEM_SEEN)
		}

		return nil
	default:
		return fmt.Errorf("invalid token type: %d", token.Type)
	}
}

func (t *TokenWriter) writeIndent() error {
	for i := 0; i < t.indentLevel; i++ {
		_, err := t.wr.Write([]byte(t.indent))
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TokenWriter) writeBytes(bss ...[]byte) error {
	for _, bs := range bss {
		_, err := t.wr.Write(bs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TokenWriter) Close() error {
	closer, isCloser := t.wr.(io.Closer)
	if isCloser {
		return closer.Close()
	}

	if t.stateStack.Peek() != TWS_END {
		return fmt.Errorf("not in end state")
	}

	return nil
}

func (t *TokenWriter) WriteObjectStart() error {
	return t.WriteToken(Token{Type: TT_OBJECT_START, Value: ""})
}
func (t *TokenWriter) WriteObjectEnd() error {
	return t.WriteToken(Token{Type: TT_OBJECT_END, Value: ""})
}
func (t *TokenWriter) WriteKey(key string) error {
	return t.WriteToken(Token{Type: TT_KEY, Value: key})
}
func (t *TokenWriter) WriteArrayStart() error {
	return t.WriteToken(Token{Type: TT_ARRAY_START, Value: ""})
}
func (t *TokenWriter) WriteArrayEnd() error {
	return t.WriteToken(Token{Type: TT_ARRAY_END, Value: ""})
}
func (t *TokenWriter) WriteStringValue(value string) error {
	return t.WriteToken(Token{Type: TT_STRING_VALUE, Value: value})
}
func (t *TokenWriter) WriteBooleanValue(value bool) error {
	if value {
		return t.WriteToken(Token{Type: TT_TRUE_VALUE, Value: ""})
	}
	return t.WriteToken(Token{Type: TT_FALSE_VALUE, Value: ""})
}
func (t *TokenWriter) WriteNumberValue(value float64) error {
	return t.WriteToken(Token{Type: TT_NUMBER_VALUE, Value: fmt.Sprintf("%e", value)})
}
func (t *TokenWriter) WriteIntegerValue(value int) error {
	return t.WriteToken(Token{Type: TT_INTEGER_VALUE, Value: strconv.Itoa(value)})
}
func (t *TokenWriter) WriteNullValue() error {
	return t.WriteToken(Token{Type: TT_NULL_VALUE, Value: ""})
}

func (t *TokenWriter) addMissingTokens(token Token) error {

	currentState := t.stateStack.Peek()
	followsValue := token.Type == TT_NUMBER_VALUE ||
		token.Type == TT_INTEGER_VALUE ||
		token.Type == TT_STRING_VALUE ||
		token.Type == TT_TRUE_VALUE ||
		token.Type == TT_FALSE_VALUE ||
		token.Type == TT_OBJECT_START ||
		token.Type == TT_ARRAY_START

	if followsValue && currentState == TWS_IN_ARRAY_ITEM_SEEN {
		err := t.WriteToken(Token{Type: TT_COMMA, Value: ""})
		if err != nil {
			return err
		}
	} else if followsValue && currentState == TWS_IN_OBJECT_KEY_SEEN {
		err := t.WriteToken(Token{Type: TT_COLON, Value: ""})
		if err != nil {
			return err
		}
	} else if token.Type == TT_KEY && currentState == TWS_IN_OBJECT_PAIR_SEEN {
		err := t.WriteToken(Token{Type: TT_COMMA, Value: ""})
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TokenWriter) WriteKeyAndStringValue(key string, value string) error {
	err := t.WriteKey(key)
	if err != nil {
		return err
	}

	return t.WriteStringValue(value)
}

func (t *TokenWriter) WriteKeyAndBooleanValue(key string, value bool) error {
	err := t.WriteKey(key)
	if err != nil {
		return err
	}

	return t.WriteBooleanValue(value)
}

func (t *TokenWriter) WriteKeyAndNumberValue(key string, value float64) error {
	err := t.WriteKey(key)
	if err != nil {
		return err
	}

	return t.WriteNumberValue(value)
}

func (t *TokenWriter) WriteKeyAndIntegerValue(key string, value int) error {
	err := t.WriteKey(key)
	if err != nil {
		return err
	}

	return t.WriteIntegerValue(value)
}

func (t *TokenWriter) WriteKeyAndNullValue(key string) error {
	err := t.WriteKey(key)
	if err != nil {
		return err
	}

	return t.WriteNullValue()
}
