package sloggergo

import (
	"github.com/godeh/sloggergo/config"
	"github.com/godeh/sloggergo/formatter"
	"github.com/godeh/sloggergo/sink"
)

// NewFromConfig creates a new logger from a JSON configuration file.
func NewFromConfig(path string) (*Logger, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return NewFromConfigStruct(cfg)
}

// NewFromConfigStruct creates a new logger from a config struct.
func NewFromConfigStruct(cfg *config.Config) (*Logger, error) {
	level := ParseLevel(cfg.Logger.Level)
	logger := New(
		WithLevel(level),
		WithCaller(cfg.Logger.AddCaller),
		WithTimeFormat(cfg.Logger.TimeFormat),
	)

	var fmt formatter.Formatter
	if cfg.Logger.Format == "json" {
		fmt = formatter.NewJSON()
	} else {
		fmt = formatter.NewText()
	}

	if cfg.Logger.Stdout.Enabled {
		var textFmt *formatter.TextFormatter
		if cfg.Logger.Format == "text" {
			textFmt = formatter.NewText()
			textFmt.DisableColors = cfg.Logger.Stdout.DisableColors
			logger.AddSink(sink.NewStdout(sink.WithFormatter(textFmt)))
		} else {
			logger.AddSink(sink.NewStdout(sink.WithFormatter(fmt)))
		}
	}

	if cfg.Logger.File.Enabled {
		fileOptions := []sink.FileOption{
			sink.WithFileFormatter(fmt),
		}
		if cfg.Logger.File.MaxSizeMB > 0 {
			fileOptions = append(
				fileOptions,
				sink.WithMaxSizeMB(cfg.Logger.File.MaxSizeMB),
				sink.WithMaxBackups(cfg.Logger.File.MaxBackups),
			)
		}
		fileSink, err := sink.NewFile(cfg.Logger.File.Path, fileOptions...)
		if err != nil {
			return nil, err
		}
		logger.AddSink(fileSink)
	}

	return logger, nil
}
