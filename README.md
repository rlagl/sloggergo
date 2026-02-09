# sloggergo

Go structured logger with zero dependencies by default.

## Features

- **Zero Dependency**: Core library has 0 external dependencies.
- **Extensible**: Interface-based usage for Sinks and Formatters.
- **Async Support**: Native asynchronous logging with buffering.
- **Observability Ready**: Custom sinks are easy to build (example provided for Elasticsearch).

## Usage

```go
package main

import (
    "github.com/godeh/sloggergo"
    "github.com/godeh/sloggergo/sink"
)

func main() {
    // Create a new logger writing to stdout
    log := sloggergo.New(
        sloggergo.WithLevel(sloggergo.InfoLevel),
        sloggergo.WithSink(sink.NewStdout()),
    )
    defer log.Close()

    log.Info("Hello world")
}
```

## JSON Configuration

```json
{
  "logger": {
    "level": "info",
    "format": "text",
    "time_format": "2006-01-02T15:04:05.999999999Z07:00",
    "add_caller": true,
    "stdout": {
      "enabled": true,
      "disable_colors": false
    },
    "file": {
      "enabled": true,
      "path": "logs/app.log",
      "max_size_mb": 10,
      "max_backups": 3
    }
  }
}
```

Schema and defaults:
- Schema: `config/schema.json`
- Defaults: `level=info`, `format=text`, `time_format=RFC3339Nano`, `add_caller=true`, `stdout.enabled=false`, `file.enabled=false`.
- Rotation: `max_size_mb>0` enables rotation. When rotating, the current file is renamed to `.1`, existing backups shift up to `.N` (`max_backups`). If `max_backups=0`, rotated files are discarded.

## Examples

Check the [examples](./examples) directory for more usage scenarios:

- [Basic Usage](./examples/basic/main.go)
- [File Logging](./examples/file/main.go)
- [Async Logging](./examples/async/main.go)
- [JSON Formatting](./examples/json/main.go)
- [Error Handling](./examples/error_handling/main.go)
- [Custom Sink (Elasticsearch)](./examples/custom_sink_elasticsearch/main.go)
- [Advanced Features (Context, Hooks, Pretty Print)](./examples/advanced_features/main.go)
