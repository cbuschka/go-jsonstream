package jsonstream

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testProducesJsonViaTokenStream(t *testing.T, indent string, expectedJson string, tokens ...Token) {
	buf := new(bytes.Buffer)
	wr := NewWriter(buf).(*tokenWriter)
	wr.SetIndent(indent)

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
	testProducesJsonViaTokenStream(t, "", expectedJson, tokens)
}

func TestWritesNull(t *testing.T) {

	tokens := Token{Type: TT_NULL_VALUE, Value: ""}
	expectedJson := "null"
	testProducesJsonViaTokenStream(t, "", expectedJson, tokens)
}

func TestWritesTrue(t *testing.T) {

	tokens := Token{Type: TT_TRUE_VALUE, Value: ""}
	expectedJson := "true"
	testProducesJsonViaTokenStream(t, "", expectedJson, tokens)
}

func TestWritesEmptyObject(t *testing.T) {

	tokens := []Token{{Type: TT_OBJECT_START, Value: ""}, {Type: TT_OBJECT_END, Value: ""}}
	expectedJson := "{}"
	testProducesJsonViaTokenStream(t, "", expectedJson, tokens...)
}

func TestWritesEmptyArray(t *testing.T) {

	tokens := []Token{{Type: TT_ARRAY_START, Value: ""}, {Type: TT_ARRAY_END, Value: ""}}
	expectedJson := "[]"
	testProducesJsonViaTokenStream(t, "", expectedJson, tokens...)
}

func TestWritesStringArrayNoIndent(t *testing.T) {

	tokens := []Token{{Type: TT_ARRAY_START, Value: ""}, {Type: TT_STRING_VALUE, Value: "value0"}, {Type: TT_COMMA, Value: ""}, {Type: TT_STRING_VALUE, Value: "value1"}, {Type: TT_ARRAY_END, Value: ""}}
	expectedJson := "[\"value0\",\"value1\"]"
	testProducesJsonViaTokenStream(t, "", expectedJson, tokens...)
}

func TestWritesObjectNoIndent(t *testing.T) {

	tokens := []Token{{Type: TT_OBJECT_START, Value: ""}, {Type: TT_KEY, Value: "key"}, {Type: TT_COLON, Value: ":"}, {Type: TT_STRING_VALUE, Value: "value"}, {Type: TT_OBJECT_END, Value: ""}}
	expectedJson := "{\"key\":\"value\"}"
	testProducesJsonViaTokenStream(t, "", expectedJson, tokens...)
}

func TestWritesSingleKeyObjectTabIndent(t *testing.T) {
	indent := "\t"
	tokens := []Token{{Type: TT_OBJECT_START, Value: ""}, {Type: TT_KEY, Value: "key"}, {Type: TT_COLON, Value: ":"}, {Type: TT_STRING_VALUE, Value: "value"}, {Type: TT_OBJECT_END, Value: ""}}
	expectedJson := "{\n\t\"key\": \"value\"\n}"
	testProducesJsonViaTokenStream(t, indent, expectedJson, tokens...)
}

func TestWritesMultiKeyObjectTabIndent(t *testing.T) {
	indent := "\t"
	tokens := []Token{{Type: TT_OBJECT_START, Value: ""},
		{Type: TT_KEY, Value: "key"}, {Type: TT_COLON, Value: ":"}, {Type: TT_STRING_VALUE, Value: "value"}, {Type: TT_COMMA, Value: ","},
		{Type: TT_KEY, Value: "key2"}, {Type: TT_COLON, Value: ":"}, {Type: TT_STRING_VALUE, Value: "value2"},
		{Type: TT_OBJECT_END, Value: ""}}
	expectedJson := "{\n\t\"key\": \"value\",\n\t\"key2\": \"value2\"\n}"
	testProducesJsonViaTokenStream(t, indent, expectedJson, tokens...)
}

func TestWritesMultiItemArrayTabIndent(t *testing.T) {
	indent := "\t"
	tokens := []Token{{Type: TT_ARRAY_START, Value: ""}, {Type: TT_STRING_VALUE, Value: "value0"}, {Type: TT_COMMA, Value: ""}, {Type: TT_STRING_VALUE, Value: "value1"}, {Type: TT_ARRAY_END, Value: ""}}
	expectedJson := "[\n\t\"value0\",\n\t\"value1\"\n]"
	testProducesJsonViaTokenStream(t, indent, expectedJson, tokens...)
}

func TestWritesMultiItemArrayWithObjectsTabIndent(t *testing.T) {
	indent := "\t"
	tokens := []Token{{Type: TT_ARRAY_START, Value: ""}, {Type: TT_STRING_VALUE, Value: "value0"}, {Type: TT_COMMA, Value: ""},
		{Type: TT_OBJECT_START, Value: ""}, {Type: TT_KEY, Value: "key1"}, {Type: TT_COLON, Value: ":"}, {Type: TT_STRING_VALUE, Value: "value1"}, {Type: TT_OBJECT_END, Value: ""}, {Type: TT_COMMA, Value: ""},
		{Type: TT_STRING_VALUE, Value: "value2"}, {Type: TT_ARRAY_END, Value: ""}}
	expectedJson := "[\n\t\"value0\",{\n\t\t\"key1\": \"value1\"\n\t},\n\t\"value2\"\n]"
	testProducesJsonViaTokenStream(t, indent, expectedJson, tokens...)
}
