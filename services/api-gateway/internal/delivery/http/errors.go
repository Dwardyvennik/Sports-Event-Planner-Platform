package http

import (
	nethttp "net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	errorInvalidRequest     = "INVALID_REQUEST"
	errorInvalidCredentials = "INVALID_CREDENTIALS"
	errorInvalidToken       = "INVALID_TOKEN"
	errorUnauthorized       = "UNAUTHORIZED"
	errorForbidden          = "FORBIDDEN"
	errorEventFull          = "EVENT_FULL"
	errorEventNotFound      = "EVENT_NOT_FOUND"
	errorUserNotFound       = "USER_NOT_FOUND"
	errorUserAlreadyExists  = "USER_ALREADY_EXISTS"
	errorServiceUnavailable = "SERVICE_UNAVAILABLE"
	errorInternal           = "INTERNAL_ERROR"
)

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeAPIError(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, errorEnvelope{Error: errorBody{Code: code, Message: message}})
}

func writeGRPCError(c *gin.Context, err error) {
	statusCode, code, message := apiErrorFromGRPC(err)
	writeAPIError(c, statusCode, code, message)
}

func apiErrorFromGRPC(err error) (int, string, string) {
	st := status.Convert(err)
	message := st.Message()
	normalized := strings.ToLower(message)

	switch st.Code() {
	case codes.InvalidArgument:
		return nethttp.StatusBadRequest, errorInvalidRequest, message
	case codes.Unauthenticated:
		if strings.Contains(normalized, "credential") {
			return nethttp.StatusUnauthorized, errorInvalidCredentials, "invalid credentials"
		}
		if strings.Contains(normalized, "token") {
			return nethttp.StatusUnauthorized, errorInvalidToken, "invalid token"
		}
		return nethttp.StatusUnauthorized, errorUnauthorized, "unauthorized"
	case codes.PermissionDenied:
		return nethttp.StatusForbidden, errorForbidden, "forbidden"
	case codes.NotFound:
		if strings.Contains(normalized, "user") {
			return nethttp.StatusNotFound, errorUserNotFound, "user not found"
		}
		return nethttp.StatusNotFound, errorEventNotFound, "event not found"
	case codes.AlreadyExists:
		return nethttp.StatusConflict, errorUserAlreadyExists, message
	case codes.FailedPrecondition:
		if strings.Contains(normalized, "full") {
			return nethttp.StatusConflict, errorEventFull, "event is full"
		}
		return nethttp.StatusConflict, errorInvalidRequest, message
	case codes.Unavailable, codes.DeadlineExceeded:
		return nethttp.StatusServiceUnavailable, errorServiceUnavailable, "service unavailable"
	default:
		return nethttp.StatusInternalServerError, errorInternal, "internal error"
	}
}
