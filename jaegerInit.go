package shared

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// InitJaegerOTEL инициализирует OpenTelemetry Jaeger TracerProvider
// используя плоский конфиг config.JaegerConfig.
func InitJaegerOTEL(cfg JaegerConfig, logger *zap.Logger) (*sdktrace.TracerProvider, error) {
	logger.Info("Initializing OpenTelemetry Jaeger exporter (via Agent UDP)",
		zap.String("agent_host", cfg.AgentHost),
		zap.String("agent_port", cfg.AgentPort),
		zap.String("service_name", cfg.ServiceName),
	)

	// Создаём Jaeger экспортер через Agent UDP
	exporter, err := jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(cfg.AgentHost),
			jaeger.WithAgentPort(cfg.AgentPort),
		),
	)
	if err != nil {
		logger.Error("Failed to create Jaeger exporter", zap.Error(err))
		return nil, err
	}

	// Настройка семплера
	var sampler sdktrace.Sampler
	if cfg.SamplerType == "const" && cfg.SamplerParam >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else {
		sampler = sdktrace.NeverSample()
	}

	// Создаём ресурс с атрибутами
	res, err := sdkresource.New(
		context.Background(),
		sdkresource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
		),
	)
	if err != nil {
		logger.Error("Failed to create OTEL resource", zap.Error(err))
		return nil, err
	}

	// Создаём TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxExportBatchSize(50),
			sdktrace.WithBatchTimeout(1*time.Second),
		),
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
	)

	// Устанавливаем глобальный TracerProvider
	otel.SetTracerProvider(tp)
	logger.Info("OpenTelemetry Jaeger TracerProvider initialized")

	return tp, nil
}
