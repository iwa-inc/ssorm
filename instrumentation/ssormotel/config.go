package ssormotel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	name = "github.com/iwa-inc/ssorm/ssormotel"
)

type config struct {
	tp     trace.TracerProvider
	tracer trace.Tracer

	attrs                []attribute.KeyValue
	enableQueryStatement bool
	statement            string
}

type Option interface {
	apply(conf *config)
}

type option func(conf *config)

func (fn option) apply(conf *config) {
	fn(conf)
}

func newConfig(opts ...Option) *config {
	tp := otel.GetTracerProvider()
	conf := &config{
		tp:     tp,
		tracer: tp.Tracer(name),
		attrs: []attribute.KeyValue{
			semconv.DBSystemKey.String("spanner"),
		},
		enableQueryStatement: false,
	}
	for _, opt := range opts {
		opt.apply(conf)
	}
	return conf
}

func WithAttributes(attrs ...attribute.KeyValue) Option {
	return option(func(conf *config) {
		conf.attrs = append(conf.attrs, attrs...)
	})
}

func WithTracerProvider(provider trace.TracerProvider) Option {
	return option(func(conf *config) {
		conf.tp = provider
	})
}

func WithQueryStatement() Option {
	return option(func(conf *config) {
		conf.enableQueryStatement = true
	})
}
