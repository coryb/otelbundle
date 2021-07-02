## OpenTelemetry utility bundle

This repo is for a few utilities that I felt were missing from current Go offerings from the [OpenTelemetry project](https://github.com/open-telemetry?language=go)

### Instrumentation

#### [`github.com/coryb/otelbundle/instrumentation/httptrace`](https://github.com/coryb/otelbundle/tree/main/instrumentation/httptrace) [![Go Reference](https://pkg.go.dev/badge/github.com/coryb/otelbundle/instrumentation/httptrace.svg)](https://pkg.go.dev/github.com/coryb/otelbundle/instrumentation/httptrace)
This package adds events to existing spans to collect low-level http trace details.

### Propagation

#### [`github.com/coryb/otelbundle/propagation/env`](https://github.com/coryb/otelbundle/tree/main/propagation/env) [![Go Reference](https://pkg.go.dev/badge/github.com/coryb/otelbundle/propagation/env.svg)](https://pkg.go.dev/github.com/coryb/otelbundle/propagation/env)

This package allows you to pass remote span context through environment variables for multiprocess tracing.