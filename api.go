package jsonstream

type Writer interface {
	StartObject() error
	EndObject() error
	Key(key string) error
	KeyAndStringValue(key string, value string) error
	KeyAndBooleanValue(key string, value bool) error
	KeyAndNumberValue(key string, value int) error
	KeyAndNull(key string) error
	StartArray() error
	EndArray() error
	StringValue(value string) error
	BooleanValue(value bool) error
	NumberValue(value int) error
	NullValue() error
	SetIndent(indent string)
	Close() error
}
