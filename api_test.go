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
		return wr.StringValue("value")
	})
}

func TestWritesNullViaWriter(t *testing.T) {
	expectedJson := "null"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		return wr.NullValue()
	})
}

func TestWritesTrueViaWriter(t *testing.T) {
	expectedJson := "true"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		return wr.BooleanValue(true)
	})
}

func TestWritesEmptyObjecViaWritert(t *testing.T) {
	expectedJson := "{}"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.StartObject()
		if err != nil {
			return err
		}
		return wr.EndObject()
	})
}

func TestWritesEmptyArrayViaWritert(t *testing.T) {
	expectedJson := "[]"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.StartArray()
		if err != nil {
			return err
		}
		return wr.EndArray()
	})
}

func TestWritesStringArrayNoIndentViaWritert(t *testing.T) {
	expectedJson := "[\"value0\",\"value1\"]"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.StartArray()
		if err != nil {
			return err
		}
		err = wr.StringValue("value0")
		if err != nil {
			return err
		}
		err = wr.StringValue("value1")
		if err != nil {
			return err
		}
		return wr.EndArray()
	})
}

func TestWritesObjectNoIndentViaWritert(t *testing.T) {
	expectedJson := "{\"key\":\"value\"}"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.StartObject()
		if err != nil {
			return err
		}
		err = wr.KeyAndStringValue("key", "value")
		if err != nil {
			return err
		}
		return wr.EndObject()
	})
}

func TestWritesObjectWithMutiplePropertiesNoIndentViaWriter(t *testing.T) {
	expectedJson := "{\"key\":\"value\",\"key2\":\"value2\"}"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.StartObject()
		if err != nil {
			return err
		}
		err = wr.KeyAndStringValue("key", "value")
		if err != nil {
			return err
		}
		err = wr.KeyAndStringValue("key2", "value2")
		if err != nil {
			return err
		}
		return wr.EndObject()
	})
}
