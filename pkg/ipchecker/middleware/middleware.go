package ipcheckermiddleware

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ry461ch/metric-collector/pkg/ipchecker"
)

// Проверка подписи пришедшего запроса
func CheckRequesterIP(ipChecker *ipchecker.IPChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ip := req.Header.Get("X-Real-IP")
			realIP := net.ParseIP(ip)
			if realIP == nil || !ipChecker.Contains(&realIP) {
				res.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(res, req)
		})
	}
}

// Проверка X-Real-IP на стороне grpc-сервера
func CheckGRPCRequesterIP(ipChecker *ipchecker.IPChecker) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.DataLoss, "missing context")
		}

		values := md.Get("X-Real-IP")
		if len(values) == 0 {
			return status.Error(codes.DataLoss, "missing ip")
		}

		realIP := net.ParseIP(values[0])
		if realIP == nil || !ipChecker.Contains(&realIP) {
			return status.Error(codes.DataLoss, "forbidden")
		}

		return handler(srv, ss)
	}
}

// Добавление X-Real-IP на стороне grpc-клиента
func SetIPGRPCClientStreamInterceptor(ip string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		newCtx := metadata.AppendToOutgoingContext(ctx, "X-Real-IP", ip)

		clientStream, err := streamer(newCtx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		return clientStream, nil
	}
}
