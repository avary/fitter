package notifier

import (
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"os"
)

type console struct {
	logger logger.Logger
	name   string
	cfg    *config.ConsoleConfig
}

func (o *console) notify(record *singleRecord) error {
	bb, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if o.cfg.OnlyResult {
		_, errOut := fmt.Fprintln(os.Stdout, string(bb))
		return errOut
	}
	o.logger.Infow("Processing done", "response", string(bb))
	return nil
}

var (
	_ Notifier = &console{}
)

func NewConsole(name string, cfg *config.ConsoleConfig) *console {
	return &console{
		logger: logger.Null,
		name:   name,
		cfg:    cfg,
	}
}

func (o *console) WithLogger(logger logger.Logger) *console {
	o.logger = logger
	return o
}

func (o *console) Inform(result *parser.ParseResult, errResult error, asArray bool) error {
	return inform(o, o.name, result, errResult, asArray, o.logger)
}
