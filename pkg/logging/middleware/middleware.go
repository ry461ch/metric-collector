package requestlogger

import (
	"time"
	// "context"
	"net/http"

	// "google.golang.org/grpc"
	// "google.golang.org/grpc/metadata"
	// "google.golang.org/grpc/status"

	"github.com/ry461ch/metric-collector/pkg/logging"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Переопределение метода Write для миддлвари логгера
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// Переопределение метода WriteHeader для логгера
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Собственно миддлваря логгера
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logging.Logger.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}

// func StreamInterceptor(ctx context.Context, req interface{}, info *grpc.StreamServerInterceptor, handler grpc.StreamHandler) (interface{}, error) {
//     var token string
//     if md, ok := metadata.FromIncomingContext(ctx); ok {
//         values := md.Get("token")
//         if len(values) > 0 {
//             token = values[0]
//         }
//     }
//     if len(token) == 0 {
//         return nil, status.Error(codes.Unauthenticated, "missing token")
//     }
//     if token != SecretToken {
//         return nil, status.Error(codes.Unauthenticated, "invalid token")
//     }
//     return handler(ctx, req)
// }
