package jsonstream

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testProducesJson(t *testing.T, expectedJson string, tokens ...Token) {
	buf := new(bytes.Buffer)
	wr := NewWriter(buf)

	err := wr.WriteTokens(tokens...)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = wr.Close()
	if err != nil {
		t.Fatal(err)
		return
	}

	json := buf.String()
	assert.Equal(t, expectedJson, json)
}

func TestWritesString(t *testing.T) {
	tokens := Token{Type: TT_STRING_VALUE, Value: "value"}
	expectedJson := "\"value\""
	testProducesJson(t, expectedJson, tokens)
}

func TestWritesNull(t *testing.T) {

	tokens := Token{Type: TT_NULL_VALUE, Value: ""}
	expectedJson := "null"
	testProducesJson(t, expectedJson, tokens)
}

func TestWritesTrue(t *testing.T) {

	tokens := Token{Type: TT_TRUE_VALUE, Value: ""}
	expectedJson := "true"
	testProducesJson(t, expectedJson, tokens)
}

func TestWritesEmptyObject(t *testing.T) {

	tokens := []Token{{Type: TT_OBJECT_START, Value: ""}, {Type: TT_OBJECT_END, Value: ""}}
	expectedJson := "{}"
	testProducesJson(t, expectedJson, tokens...)
}

func TestWritesEmptyArray(t *testing.T) {

	tokens := []Token{{Type: TT_ARRAY_START, Value: ""}, {Type: TT_ARRAY_END, Value: ""}}
	expectedJson := "[]"
	testProducesJson(t, expectedJson, tokens...)
}

func TestWritesStringArray(t *testing.T) {

	tokens := []Token{{Type: TT_ARRAY_START, Value: ""}, {Type: TT_STRING_VALUE, Value: "value0"}, {Type: TT_COMMA, Value: ""}, {Type: TT_STRING_VALUE, Value: "value1"}, {Type: TT_ARRAY_END, Value: ""}}
	expectedJson := "[\"value0\",\"value1\"]"
	testProducesJson(t, expectedJson, tokens...)
}
