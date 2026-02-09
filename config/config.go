package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the complete logger configuration.
type Config struct {
	Logger LoggerConfig `json:"logger"`
}

// LoggerConfig contains all logger settings.
type LoggerConfig struct {
	// Level is the minimum log level (debug, info, warn, error, fatal)
	Level string `json:"level"`

	// Format is the output format (json, text)
	Format string `json:"format"`

	// TimeFormat is the time format for logs (default: RFC3339Nano)
	TimeFormat string `json:"time_format"`

	// AddCaller enables caller information
	AddCaller bool `json:"add_caller"`

	// Stdout configuration
	Stdout StdoutConfig `json:"stdout"`

	// File configuration
	File FileConfig `json:"file"`
}

// StdoutConfig configures stdout output.
type StdoutConfig struct {
	Enabled       bool `json:"enabled"`
	DisableColors bool `json:"disable_colors"`
}

// FileConfig configures file output.
type FileConfig struct {
	Enabled    bool   `json:"enabled"`
	Path       string `json:"path"`
	MaxSizeMB  int    `json:"max_size_mb"`
	MaxBackups int    `json:"max_backups"`
}

// Load reads and parses a configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Set defaults
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.Format == "" {
		cfg.Logger.Format = "text"
	}
	if cfg.Logger.TimeFormat == "" {
		cfg.Logger.TimeFormat = time.RFC3339Nano
	}
	if !hasLoggerField(raw, "add_caller") {
		cfg.Logger.AddCaller = true
	}

	return &cfg, nil
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	// Validate level
	level := c.Logger.Level
	if level != "debug" && level != "info" && level != "warn" && level != "error" && level != "fatal" {
		return fmt.Errorf("invalid log level: %s", level)
	}

	// Validate format
	format := c.Logger.Format
	if format != "json" && format != "text" {
		return fmt.Errorf("invalid format: %s", format)
	}

	// Validate file path if enabled
	if c.Logger.File.Enabled && c.Logger.File.Path == "" {
		return fmt.Errorf("file path is required when file output is enabled")
	}
	if c.Logger.File.MaxSizeMB < 0 {
		return fmt.Errorf("max_size_mb cannot be negative")
	}
	if c.Logger.File.MaxBackups < 0 {
		return fmt.Errorf("max_backups cannot be negative")
	}

	return nil
}

func hasLoggerField(raw map[string]json.RawMessage, field string) bool {
	rawLogger, ok := raw["logger"]
	if !ok {
		return false
	}

	var loggerFields map[string]json.RawMessage
	if err := json.Unmarshal(rawLogger, &loggerFields); err != nil {
		return false
	}

	_, ok = loggerFields[field]
	return ok
}
