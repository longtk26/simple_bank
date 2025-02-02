package gapi

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(
	ctx context.Context, 
	req any, 
	info *grpc.UnaryServerInfo, 
	handler grpc.UnaryHandler,
) (resp any, err error) {
	startTime := time.Now()
	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()	
	}

	logger := log.Info()

	if err != nil {
		logger = log.Error().Err(err)
	}
	
	logger.Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Dur("duration", duration).
		Msg("Received gRPC request")

	return result, err
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body []byte
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	r.Body = data
	return r.ResponseWriter.Write(data)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		recorder := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode: http.StatusOK,
		}
		handler.ServeHTTP(recorder, req)
		duration := time.Since(startTime)
		logger := log.Info()

		if recorder.StatusCode != http.StatusOK {
			logger = log.Error().Bytes("response_body", recorder.Body)
		}

		logger.Str("protocol", "http").
		Str("method", req.Method).
		Str("path", req.RequestURI).
		Int("status_code", recorder.StatusCode).
		Str("status_text", http.StatusText(recorder.StatusCode)).
		Dur("duration", duration).
		Msg("Received http request")

	})
}