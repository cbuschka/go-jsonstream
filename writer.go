package jsonstream

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

type tokenWriter struct {
	wr          io.Writer
	indent      string
	indentLevel int
	stateStack  tokenWriterStateStack
}

func NewWriter(wr io.Writer) Writer {
	return Writer(&tokenWriter{wr: wr, indent: "", indentLevel: 0, stateStack: tokenWriterStateStack{TWS_INITIAL}})
}

func (t *tokenWriter) SetIndent(indent string) {
	t.indent = indent
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

func (t *tokenWriter) checkTokenAllowed(currTokenType TokenType, allowedStates ...tokenWriterState) error {
	currState := t.stateStack.Peek()
	for _, allowedState := range allowedStates {
		if allowedState == currState {
			return nil
		}
	}

	return fmt.Errorf("%s not allowed in %s", currTokenType.Name(), currState.Name())
}

func (t *tokenWriter) WriteToken(token Token) error {

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
	default:
		return fmt.Errorf("invalid token type: %d", token.Type)
	}
}

func (t *tokenWriter) writeIndent() error {
	for i := 0; i < t.indentLevel; i++ {
		_, err := t.wr.Write([]byte(t.indent))
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *tokenWriter) writeBytes(bss ...[]byte) error {
	for _, bs := range bss {
		_, err := t.wr.Write(bs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *tokenWriter) Close() error {
	closer, isCloser := t.wr.(io.Closer)
	if isCloser {
		return closer.Close()
	}

	return nil
}

func (t *tokenWriter) StartObject() error {
	return t.WriteToken(Token{Type: TT_OBJECT_START, Value: ""})
}
func (t *tokenWriter) EndObject() error {
	return t.WriteToken(Token{Type: TT_OBJECT_END, Value: ""})
}
func (t *tokenWriter) Key(key string) error {
	return t.WriteToken(Token{Type: TT_KEY, Value: key})
}
func (t *tokenWriter) StartArray() error {
	return t.WriteToken(Token{Type: TT_ARRAY_START, Value: ""})
}
func (t *tokenWriter) EndArray() error {
	return t.WriteToken(Token{Type: TT_ARRAY_END, Value: ""})
}
func (t *tokenWriter) String(value string) error {
	return t.WriteToken(Token{Type: TT_STRING_VALUE, Value: value})
}
func (t *tokenWriter) Boolean(value bool) error {
	if value {
		return t.WriteToken(Token{Type: TT_TRUE_VALUE, Value: ""})
	}
	return t.WriteToken(Token{Type: TT_FALSE_VALUE, Value: ""})
}
func (t *tokenWriter) Number(value int) error {
	return t.WriteToken(Token{Type: TT_NUMBER_VALUE, Value: strconv.Itoa(value)})
}
func (t *tokenWriter) Null() error {
	return t.WriteToken(Token{Type: TT_NULL_VALUE, Value: ""})
}

func (t *tokenWriter) addMissingTokens(token Token) error {

	currentState := t.stateStack.Peek()
	followsValue := token.Type == TT_NUMBER_VALUE ||
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
	}

	return nil
}
