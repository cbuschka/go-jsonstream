package jsonstream

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testProducesJsonViaWriter(t *testing.T, expectedJson string, writeFunc func(wr Writer) error) {
	buf := new(bytes.Buffer)
	wr := NewWriter(buf).(*tokenWriter)
	wr.SetIndent("")

	err := writeFunc(wr)
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

func TestWritesStringViaWriter(t *testing.T) {
	expectedJson := "\"value\""
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		return wr.String("value")
	})
}

func TestWritesNullViaWriter(t *testing.T) {
	expectedJson := "null"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		return wr.Null()
	})
}

func TestWritesTrueViaWriter(t *testing.T) {
	expectedJson := "true"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		return wr.Boolean(true)
	})
}

func TestWritesEmptyObjecViaWritert(t *testing.T) {
	expectedJson := "{}"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		_ = wr.StartObject()
		return wr.EndObject()
	})
}

func TestWritesEmptyArrayViaWritert(t *testing.T) {
	expectedJson := "[]"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		_ = wr.StartArray()
		return wr.EndArray()
	})
}

func TestWritesStringArrayNoIndentViaWritert(t *testing.T) {
	expectedJson := "[\"value0\",\"value1\"]"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		_ = wr.StartArray()
		_ = wr.String("value0")
		_ = wr.String("value1")
		return wr.EndArray()
	})
}

func TestWritesObjectNoIndentViaWritert(t *testing.T) {
	expectedJson := "{\"key\":\"value\"}"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		_ = wr.StartObject()
		_ = wr.KeyAndStringValue("key", "value")
		return wr.EndObject()
	})
}
