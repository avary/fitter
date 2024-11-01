package utils

import (
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/expr-lang/expr"
)

const (
	fitterResultJsonRef           = "fResJson"
	fitterResultRef               = "fRes"
	fitterIndexRef                = "fIndex"
	fitterResultRaw               = "fResRaw"
	fitterNewLinePlaceholderKey   = "FNewLine"
	fitterNewLinePlaceholderValue = "$__FLINE__$"
)

var (
	defEnv = map[string]interface{}{
		fitterNewLinePlaceholderKey: fitterNewLinePlaceholderValue,
		"FNull":                     builder.NullValue,
		"FNil":                      nil,
		"isNull": func(value interface{}) bool {
			return builder.NullValue == value
		},
	}
)

func extendEnv(env map[string]interface{}, result builder.Interfacable, index *uint32) map[string]interface{} {
	kv := make(map[string]interface{})

	for k, v := range env {
		kv[k] = v
	}

	if result != nil {
		kv[fitterResultRaw] = result.Raw()
		kv[fitterResultRef] = result.ToInterface()
		kv[fitterResultJsonRef] = result.ToJson()
	}
	if index != nil {
		kv[fitterIndexRef] = *index
	}

	return kv
}

func ProcessExpression(expression string, result builder.Interfacable, index *uint32, input builder.Interfacable) (builder.Interfacable, error) {
	env := extendEnv(defEnv, result, index)
	program, err := expr.Compile(Format(expression, result, index, input), expr.Env(env))
	if err != nil {
		return nil, err
	}

	out, err := expr.Run(program, env)
	if err != nil {
		return nil, err
	}

	bb, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return builder.ToJsonable(bb), nil
}
