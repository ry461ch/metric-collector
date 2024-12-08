package requestlogger

import (
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

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

	loggingStreamServer struct {
		grpc.ServerStream
		size int
	}
)

// Переопределение метода RecvMsg для grpc миддлвари логгера
func (lss *loggingStreamServer) RecvMsg(req interface{}) error {
	err := lss.ServerStream.RecvMsg(req)
	if err != nil {
		return err
	}
	if msg, ok := req.(proto.Message); ok {
		lss.size += proto.Size(msg)
	}
	return nil
}

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
			"type", "http",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}

// Interceptor логирования запросов grpc
func LoggingStreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()

	loggingStreamServer := &loggingStreamServer{
		ServerStream: ss,
		size:         0,
	}

	err := handler(srv, loggingStreamServer)
	if err != nil {
		logging.Logger.Errorln(err)
	}

	duration := time.Since(start)
	logging.Logger.Infoln(
		"type", "grpc",
		"method", info.FullMethod,
		"duration", duration,
		"size", loggingStreamServer.size,
	)
	return err
}
