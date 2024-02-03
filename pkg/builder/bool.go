package builder

import (
	"encoding/json"
	"fmt"
)

type boolField struct {
	value bool
}

var (
	_ Jsonable = &boolField{}
)

func Bool(value bool) *boolField {
	return &boolField{
		value: value,
	}
}

func (s *boolField) IsEmpty() bool {
	return false
}

func (s *boolField) ToJson() string {
	return fmt.Sprintf(`%v`, s.value)
}

func (s *boolField) Raw() json.RawMessage {
	return toRaw(s.value)
}
