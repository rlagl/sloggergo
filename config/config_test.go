package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{"logger":{}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Logger.Level != "info" {
		t.Fatalf("expected level=info, got %q", cfg.Logger.Level)
	}
	if cfg.Logger.Format != "text" {
		t.Fatalf("expected format=text, got %q", cfg.Logger.Format)
	}
	if cfg.Logger.TimeFormat != time.RFC3339Nano {
		t.Fatalf("expected time_format=%q, got %q", time.RFC3339Nano, cfg.Logger.TimeFormat)
	}
	if cfg.Logger.AddCaller != true {
		t.Fatalf("expected add_caller default true, got %v", cfg.Logger.AddCaller)
	}
}

func TestLoadAddCallerFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{"logger":{"add_caller":false}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Logger.AddCaller != false {
		t.Fatalf("expected add_caller=false, got %v", cfg.Logger.AddCaller)
	}
}
