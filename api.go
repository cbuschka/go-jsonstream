package jsonstream

import (
	"github.com/cbuschka/go-jsonstream/internal"
	"io"
)

type Writer interface {
	WriteObjectStart() error
	WriteObjectEnd() error
	WriteKey(key string) error
	WriteKeyAndStringValue(key string, value string) error
	WriteKeyAndBooleanValue(key string, value bool) error
	WriteKeyAndNumberValue(key string, value float64) error
	WriteKeyAndIntegerValue(key string, value int) error
	WriteKeyAndNullValue(key string) error
	WriteArrayStart() error
	WriteArrayEnd() error
	WriteStringValue(value string) error
	WriteBooleanValue(value bool) error
	WriteNumberValue(value float64) error
	WriteIntegerValue(value int) error
	WriteNullValue() error
	SetIndent(indent string)
	Close() error
}

func NewWriter(wr io.Writer) Writer {
	return Writer(internal.NewTokenWriter(wr))
}
