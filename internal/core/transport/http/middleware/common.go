package core_http_middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	core_http_response "github.com/qandoni/debatesApp/internal/core/transport/http/response"

	"go.uber.org/zap"
)

const (
	requestIDHeader = "X-Request-ID"
	loggerKey       = "logger"
)

func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx.Writer.Header().Set(requestIDHeader, requestID)
		ctx.Request.Header.Set(requestIDHeader, requestID)

		ctx.Set(requestIDHeader, requestID)
		ctx.Next()
	}
}

func Logger(log *core_logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)

		requestLogger := log.With(
			zap.String("requestID", requestID),
			zap.String("url", c.Request.URL.String()),
		)

		c.Set(loggerKey, requestLogger)

		ctx := core_logger.ToContext(
			c.Request.Context(),
			requestLogger,
		)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := core_logger.FromContext(c.Request.Context())

		before := time.Now()

		log.Debug(
			">>> incoming HTTP request",
			zap.String("http_method", c.Request.Method),
			zap.Time("time", before.UTC()),
		)

		c.Next()

		log.Debug(
			"<<< done HTTP request",
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("latence", time.Since(before)),
		)
	}
}

type TokenParser interface {
	ParseAccessToken(
		token string,
	) (core_auth.AuthInfo, error)
}

func JWT(
	parser TokenParser,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(core_errors.ErrUnauthorized).SetMeta("missing authorization header")
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			c.Error(core_errors.ErrUnauthorized).SetMeta("invalid authorizaiton header")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, prefix)

		authInfo, err := parser.ParseAccessToken(
			token,
		)
		if err != nil {
			c.Error(err).SetMeta("invalid token")
			c.Abort()
			return
		}
		ctx := core_auth.WithAuthInfo(
			c.Request.Context(),
			authInfo,
		)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last()

		log := core_logger.FromContext(c.Request.Context())

		var (
			statusCode int
			logFunc    func(string, ...zap.Field)
		)

		switch {
		case errors.Is(err.Err, core_errors.ErrUnauthorized):
			statusCode = http.StatusUnauthorized
			logFunc = log.Warn
		case errors.Is(err.Err, core_errors.ErrInvalidArgument):
			statusCode = http.StatusBadRequest
			logFunc = log.Warn
		case errors.Is(err.Err, core_errors.ErrConflict):
			statusCode = http.StatusConflict
			logFunc = log.Warn
		case errors.Is(err.Err, core_errors.ErrNotFound):
			statusCode = http.StatusNotFound
			logFunc = log.Debug
		case errors.Is(err.Err, core_errors.ErrAccessForbidden):
			statusCode = http.StatusForbidden
			logFunc = log.Debug
		case errors.Is(err.Err, core_errors.ErrEmptyFile):
			statusCode = http.StatusBadRequest
			logFunc = log.Debug

		default:
			statusCode = http.StatusInternalServerError
			logFunc = log.Error
		}
		logFunc(err.Error(), zap.Error(err.Err))

		message := "request failed"
		if err.Meta != nil {
			message = err.Meta.(string)
		}
		c.AbortWithStatusJSON(statusCode, core_http_response.ErrorResponse{
			Error:   err.Err.Error(),
			Message: message,
		})
	}
}
