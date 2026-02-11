package shared

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

func OTELHTTPMiddleware(tracer trace.Tracer, logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		propagator := propagation.TraceContext{}
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Создаём span для всего HTTP запроса
		ctx, span := tracer.Start(ctx, "http.request",
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
			),
		)
		defer func() {
			span.End()
			logger.Info("finished HTTP request span",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
			)
		}()

		// Логируем начало запроса и traceID
		traceID := span.SpanContext().TraceID().String()
		logger.Info("HTTP request started",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("trace_id", traceID),
		)

		// Прокидываем контекст дальше
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
