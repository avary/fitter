package builder

import (
	"fmt"
	"strings"
)

type stringField struct {
	value string
}

func String(value string) *stringField {
	return &stringField{
		value: value,
	}
}

func (s *stringField) ToJson() string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(s.value, `"`, `\"`))
}