# sloggergo

Go structured logger with zero dependencies by default.

## Features

- **Zero Dependency**: Core library has 0 external dependencies.
- **Extensible**: Interface-based usage for Sinks and Formatters.
- **Async Support**: Native asynchronous logging with buffering.
- **Observability Ready**: Built-in support (via standard library HTTP) for Elasticsearch, Loki, and Datadog.

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

## Examples

Check the [examples](./examples) directory for more usage scenarios:

- [Basic Usage](./examples/basic/main.go)
- [File Logging](./examples/file/main.go)
- [Async Logging](./examples/async/main.go)
- [JSON Formatting](./examples/json/main.go)
- [Error Handling](./examples/error_handling/main.go)
- [Custom Sink (Elasticsearch)](./examples/custom_sink_elasticsearch/main.go)
- [Advanced Features (Context, Hooks, Pretty Print)](./examples/advanced_features/main.go)