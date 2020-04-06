## dfuse Tracing Library

This repository contains all common stuff around trace(s) handling across our
various services

### Philosophy

The package provides a quick setup function with sensible defaults that can be used across
all our micro-services in one shot.

The `SetupTracing` function make sensible decisions to setup tracing exporters based
on the environment. If in production, registers the `StackDriver` exporter
with a probability sampler of 1/4. In development, registers exporters based on environment
variables `TRACING_ZAP_EXPORTER` (zap exporter) and `TRACING_ZIPKIN_EXPORTER=zipkinURL` for
ZipKin exporter.

For easier customization in package, we also exposes all `Register*` functions so it's possible
to easily customize the behavior.
