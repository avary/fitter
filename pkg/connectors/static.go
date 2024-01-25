package connectors

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/utils"
)

type staticConnector struct {
	cfg    *config.StaticConnectorConfig
	logger logger.Logger
}

func (j *staticConnector) Get(parsedValue builder.Jsonable, index *uint32) ([]byte, error) {
	if len(j.cfg.Raw) != 0 {
		return []byte(utils.Format(string(j.cfg.Raw), parsedValue, index)), nil
	}
	return []byte(utils.Format(j.cfg.Value, parsedValue, index)), nil
}

func NewStatic(cfg *config.StaticConnectorConfig) *staticConnector {
	return &staticConnector{
		cfg:    cfg,
		logger: logger.Null,
	}
}

func (j *staticConnector) WithLogger(logger logger.Logger) *staticConnector {
	j.logger = logger
	return j
}
