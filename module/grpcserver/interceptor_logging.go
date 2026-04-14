package grpcserver

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
)

// LoggingInterceptor returns a grpc.UnaryServerInterceptor that logs incoming GRPC request and response.
// It extracts the requester's peer address from the gRPC context and includes it in log entries.
// Logs are emitted at the start and finish of each call for debugging and tracing purposes.
func LoggingInterceptor(log zerolog.Logger) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(
		InterceptorLogger(log),
		logging.WithLevels(statusCodeToLogLevel),
		logging.WithFieldsFromContext(extractPeerFields),
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	)
}

// StreamLoggingInterceptor returns a grpc.StreamServerInterceptor that logs incoming streaming GRPC requests.
// It extracts the requester's peer address from the gRPC context and includes it in log entries.
// Logs are emitted at the start and finish of each streaming call for debugging and tracing purposes.
func StreamLoggingInterceptor(log zerolog.Logger) grpc.StreamServerInterceptor {
	return logging.StreamServerInterceptor(
		InterceptorLogger(log),
		logging.WithLevels(statusCodeToLogLevel),
		logging.WithFieldsFromContext(extractPeerFields),
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	)
}

// extractPeerFields extracts peer information from the gRPC context and returns it as logging fields.
// This includes the client's remote address for request tracing and debugging.
func extractPeerFields(ctx context.Context) logging.Fields {
	if p, ok := peer.FromContext(ctx); ok {
		return logging.Fields{"peer.address", p.Addr.String()}
	}
	return nil
}

// InterceptorLogger adapts a zerolog.Logger to interceptor's logging.Logger
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		l := l.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			l.Debug().Msg(msg)
		case logging.LevelInfo:
			l.Info().Msg(msg)
		case logging.LevelWarn:
			l.Warn().Msg(msg)
		case logging.LevelError:
			l.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

// statusCodeToLogLevel converts a grpc status.Code to the appropriate logging.Level
func statusCodeToLogLevel(c codes.Code) logging.Level {
	switch c {
	case codes.OK:
		// log successful returns as Debug to avoid excessive logging in info mode
		return logging.LevelDebug
	case codes.DeadlineExceeded, codes.ResourceExhausted, codes.OutOfRange:
		// these are common, map to info
		return logging.LevelInfo
	default:
		return logging.DefaultServerCodeToLevel(c)
	}
}
