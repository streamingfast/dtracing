# StreamingFast Tracing Library

[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/dtracing)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
This repository contains all common stuff around trace(s) handling across our
various services


## Philosophy

The package provides a quick setup function with sensible defaults that can be used across
all our micro-services in one shot.

The `SetupTracing` function make sensible decisions to setup tracing exporters based
on the environment. If in production, registers the `StackDriver` exporter
with a probability sampler of 1/4. In development, registers exporters based on environment
variables `TRACING_ZAP_EXPORTER` (zap exporter) and `TRACING_ZIPKIN_EXPORTER=zipkinURL` for
ZipKin exporter.

For easier customization in package, we also exposes all `Register*` functions so it's possible
to easily customize the behavior.


## Contributing

**Issues and PR in this repo related strictly to the dtracing library.**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.


## License

[Apache 2.0](LICENSE)
