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
		value: strings.TrimSpace(value),
	}
}

func (s *stringField) ToJson() string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(strings.ReplaceAll(s.value, `"`, `\"`), `'`, `\'`))
}
