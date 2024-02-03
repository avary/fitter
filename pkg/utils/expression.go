package utils

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/expr-lang/expr"
)

const (
	fitterResultJsonRef = "fResJson"
	fitterResultRef     = "fRes"
	fitterIndexRef      = "fIndex"
)

var (
	defEnv = map[string]interface{}{}
)

func extendEnv(env map[string]interface{}, result builder.Jsonable, index *uint32) map[string]interface{} {
	kv := make(map[string]interface{})

	for k, v := range env {
		kv[k] = v
	}

	if result != nil {
		res, err := expr.Eval("fromJSON(val)", map[string]interface{}{
			"val": string(result.Raw()),
		})
		if err != nil {
			res = builder.NullValue.Raw()
		}
		kv[fitterResultRef] = res
		kv[fitterResultJsonRef] = result.ToJson()
	}
	if index != nil {
		kv[fitterIndexRef] = *index
	}

	return kv
}

func ProcessExpression(expression string, result builder.Jsonable, index *uint32, input builder.Jsonable) (interface{}, error) {
	env := extendEnv(defEnv, result, index)

	program, err := expr.Compile(Format(expression, result, index, input), expr.Env(env))
	if err != nil {
		return false, err
	}

	out, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}

	return out, nil
}
