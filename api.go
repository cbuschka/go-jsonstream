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
	String(value string) error
	Boolean(value bool) error
	Number(value int) error
	Null() error
	SetIndent(indent string)
	Close() error
}
