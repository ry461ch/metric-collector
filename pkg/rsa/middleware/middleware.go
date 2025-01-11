package rsamiddleware

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/ry461ch/metric-collector/pkg/rsa"
	pb "github.com/ry461ch/metric-collector/pkg/rsa/encrypted"
)

// Проверка подписи пришедшего запроса
func DecryptRequest(decrypter *rsa.RsaDecrypter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			reqBody := buf.Bytes()
			reqDecrypted, err := decrypter.Decrypt(reqBody)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(reqDecrypted))
			next.ServeHTTP(w, r)
		})
	}
}

type decrypterServerStream struct {
	grpc.ServerStream
	decrypter *rsa.RsaDecrypter
}

// Расшифровка полученных данных в grpc interceptore
func (dss *decrypterServerStream) RecvMsg(req interface{}) error {
	msg := &pb.EncryptedObject{}
	err := dss.ServerStream.RecvMsg(msg)
	if err != nil {
		return err
	}

	reqDecrypted, err := dss.decrypter.Decrypt(msg.Data)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "can't parse input data: %v", err)
	}

	return proto.Unmarshal(reqDecrypted, req.(proto.Message))
}

// Interceptor расшифровки сообщений на стороне сервера
func DecryptStreamServerInterceptor(decrypter *rsa.RsaDecrypter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrappedStream := &decrypterServerStream{
			ServerStream: ss,
			decrypter:    decrypter,
		}

		return handler(srv, wrappedStream)
	}
}

type encryptedClientStream struct {
	grpc.ClientStream
	encrypter *rsa.RsaEncrypter
}

// Шифровка сообщения на стороне grpc клиента
func (ecs *encryptedClientStream) SendMsg(req interface{}) error {
	if msg, ok := req.(proto.Message); ok {
		data, err := proto.Marshal(msg)
		if err != nil {
			return err
		}

		encryptedData, err := ecs.encrypter.Encrypt(data)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to encrypt message: %v", err)
		}

		msgBytes := &pb.EncryptedObject{
			Data: encryptedData,
		}

		return ecs.ClientStream.SendMsg(msgBytes)
	}
	return status.Errorf(codes.InvalidArgument, "message is not protobuf")
}

// Interceptor шифровки сообщений на стороне клиента
func EncryptStreamClientInterceptor(encrypter *rsa.RsaEncrypter) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		return &encryptedClientStream{
			ClientStream: clientStream,
			encrypter:    encrypter,
		}, nil
	}
}
