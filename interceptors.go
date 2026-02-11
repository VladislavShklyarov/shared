package shared

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TracingUnaryClientInterceptor — gRPC client interceptor для OpenTelemetry
func TracingUnaryClientInterceptor(logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		// Получаем текущий span из ctx
		span := trace.SpanFromContext(ctx)
		if span == nil || !span.SpanContext().IsValid() {
			logger.Debug("No active span in context, calling gRPC without propagation")
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// Берём outgoing metadata или создаём новую
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		// Создаём propagator и инжектим контекст span в metadata
		propagator := propagation.TraceContext{}
		propagator.Inject(ctx, metadataTextMapCarrier(md))

		// Создаём новый контекст с metadata
		ctx = metadata.NewOutgoingContext(ctx, md)

		logger.Debug("Injected OTEL trace context into gRPC metadata",
			zap.String("method", method),
		)

		// Вызываем gRPC
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// metadataTextMapCarrier конвертирует gRPC metadata в TextMapCarrier для OTEL
type metadataTextMapCarrier metadata.MD

func (m metadataTextMapCarrier) Get(key string) string {
	values := m[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (m metadataTextMapCarrier) Set(key, value string) {
	m[key] = []string{value}
}

func (m metadataTextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TracingUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("TracingUnaryServerInterceptor")

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, _ := metadata.FromIncomingContext(ctx)
		propagator := propagation.TraceContext{}
		ctx = propagator.Extract(ctx, metadataToCarrier(md))

		logger.Info("gRPC interceptor: context extracted", zap.String("method", info.FullMethod))

		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		return handler(ctx, req)
	}
}

func metadataToCarrier(md metadata.MD) propagation.TextMapCarrier {
	carrier := make(map[string]string)
	for k, vals := range md {
		if len(vals) > 0 {
			carrier[k] = vals[0]
		}
	}
	return propagation.MapCarrier(carrier)
}
