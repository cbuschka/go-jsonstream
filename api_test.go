package jsonstream

import (
	"bytes"
	"fmt"
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
		return wr.WriteStringValue("value")
	})
}

func TestWritesNullViaWriter(t *testing.T) {
	expectedJson := "null"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		return wr.WriteNullValue()
	})
}

func TestWritesTrueViaWriter(t *testing.T) {
	expectedJson := "true"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		return wr.WriteBooleanValue(true)
	})
}

func TestWritesEmptyObjecViaWritert(t *testing.T) {
	expectedJson := "{}"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.WriteObjectStart()
		if err != nil {
			return err
		}
		return wr.WriteObjectEnd()
	})
}

func TestWritesEmptyArrayViaWritert(t *testing.T) {
	expectedJson := "[]"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.WriteArrayStart()
		if err != nil {
			return err
		}
		return wr.WriteArrayEnd()
	})
}

func TestWritesStringArrayNoIndentViaWritert(t *testing.T) {
	expectedJson := "[\"value0\",\"value1\"]"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.WriteArrayStart()
		if err != nil {
			return err
		}
		err = wr.WriteStringValue("value0")
		if err != nil {
			return err
		}
		err = wr.WriteStringValue("value1")
		if err != nil {
			return err
		}
		return wr.WriteArrayEnd()
	})
}

func TestWritesObjectNoIndentViaWritert(t *testing.T) {
	expectedJson := "{\"key\":\"value\"}"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.WriteObjectStart()
		if err != nil {
			return err
		}
		err = wr.WriteKeyAndStringValue("key", "value")
		if err != nil {
			return err
		}
		return wr.WriteObjectEnd()
	})
}

func TestWritesObjectWithMutiplePropertiesNoIndentViaWriter(t *testing.T) {
	expectedJson := "{\"key\":\"value\",\"key2\":\"value2\"}"
	testProducesJsonViaWriter(t, expectedJson, func(wr Writer) error {
		err := wr.WriteObjectStart()
		if err != nil {
			return err
		}
		err = wr.WriteKeyAndStringValue("key", "value")
		if err != nil {
			return err
		}
		err = wr.WriteKeyAndStringValue("key2", "value2")
		if err != nil {
			return err
		}
		return wr.WriteObjectEnd()
	})
}

func TestFailsOnCloseIfNotInEndState(t *testing.T) {
	buf := new(bytes.Buffer)
	wr := NewWriter(buf).(*tokenWriter)
	wr.SetIndent("")

	err := wr.WriteArrayStart()
	if err != nil {
		t.Fatal(err)
		return
	}

	err = wr.Close()
	assert.Equal(t, fmt.Errorf("not in end state"), err)
}
