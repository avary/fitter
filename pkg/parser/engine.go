package parser

import (
	"errors"
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/plugins/store"
	"github.com/PxyUp/fitter/pkg/references"
	"github.com/PxyUp/fitter/pkg/utils"
)

var (
	errMissingModelConfig = errors.New("missing model config")
	errInvalid            = errors.New("invalid engine")

	_          Engine = &engine{}
	nullEngine Engine = &null{}
)

type Engine interface {
	Get(model *config.Model, parsedValue builder.Jsonable, index *uint32) (*ParseResult, error)
}

type engine struct {
	connector connectors.Connector
	parser    Factory
	logger    logger.Logger
}

type null struct {
}

func (n *null) Get(model *config.Model, parsedValue builder.Jsonable, index *uint32) (*ParseResult, error) {
	return nil, errInvalid
}

func (e *engine) Get(model *config.Model, parsedValue builder.Jsonable, index *uint32) (*ParseResult, error) {
	if model == nil {
		return nil, errMissingModelConfig
	}
	body, err := e.connector.Get(parsedValue, index)
	if err != nil {
		e.logger.Errorw("connector return error during fetch data", "error", err.Error())
		return nil, err
	}
	e.logger.Debugw("connector answer", "content", string(body))
	return e.parser(body, e.logger).Parse(model)
}

func NewEngine(cfg *config.ConnectorConfig, logger logger.Logger) Engine {
	if cfg == nil {
		return nullEngine
	}

	var connector connectors.Connector
	if cfg.StaticConfig != nil {
		connector = connectors.NewStatic(cfg.StaticConfig).WithLogger(logger.With("connector", "static"))
	}
	if cfg.ServerConfig != nil {
		connector = connectors.NewAPI(cfg.Url, cfg.ServerConfig, nil).WithLogger(logger.With("connector", "server"))
	}
	if cfg.BrowserConfig != nil {
		connector = connectors.NewBrowser(cfg.Url, cfg.BrowserConfig).WithLogger(logger.With("connector", "browser"))
	}
	if cfg.PluginConnectorConfig != nil {
		connector = store.Store.GetConnectorPlugin(cfg.PluginConnectorConfig.Name, cfg.PluginConnectorConfig, logger.With("connector", cfg.PluginConnectorConfig.Name))
	}
	if cfg.ReferenceConfig != nil {
		logger.Debugw("get value from reference store", "type", string(cfg.ResponseType), "name", cfg.ReferenceConfig.Name)
		if cfg.ResponseType == config.Json {
			connector = connectors.NewStatic(&config.StaticConnectorConfig{
				Value: references.Get(cfg.ReferenceConfig.Name).ToJson(),
			})
		}
		if cfg.ResponseType == config.XPath || cfg.ResponseType == config.HTML {
			var htmlValue string
			rawValue, ok := references.Get(cfg.ReferenceConfig.Name).Raw().(string)
			if ok {
				htmlValue = rawValue
			} else {
				htmlValue = "<html></html>"
			}
			connector = connectors.NewStatic(&config.StaticConnectorConfig{
				Value: htmlValue,
			})
		}

	}
	if cfg.IntSequenceConfig != nil {
		genSlice := utils.SafeNewSliceGenerator(cfg.IntSequenceConfig.Start, cfg.IntSequenceConfig.End, cfg.IntSequenceConfig.Step)
		logger.Debugw("generated slice", "length", fmt.Sprintf("%d", len(genSlice)), "start", fmt.Sprintf("%d", cfg.IntSequenceConfig.Start), "end", fmt.Sprintf("%d", cfg.IntSequenceConfig.End), "step", fmt.Sprintf("%d", cfg.IntSequenceConfig.Step))
		jsonArr := make([]builder.Jsonable, len(genSlice))
		for i, v := range genSlice {
			jsonArr[i] = builder.Int(v)
		}
		connector = connectors.NewStatic(&config.StaticConnectorConfig{
			Value: builder.Array(jsonArr).ToJson(),
		})
	}

	var parserFactory Factory
	if cfg.ResponseType == config.Json {
		parserFactory = JsonFactory
	}
	if cfg.ResponseType == config.HTML {
		parserFactory = HTMLFactory
	}
	if cfg.ResponseType == config.XPath {
		parserFactory = XPathFactory
	}

	if connector == nil || parserFactory == nil {
		return nullEngine
	}

	connector = connectors.WithAttempts(connector, cfg.Attempts)

	return &engine{
		connector: connector,
		parser:    parserFactory,
		logger:    logger,
	}
}
