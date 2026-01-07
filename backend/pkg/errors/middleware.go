package errors

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error ErrorInfo `json:"error"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	SubCode string `json:"sub_code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// ErrorHandlerMiddleware handles errors and formats them consistently
func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Try to extract AppError
			var appErr *AppError
			if e, ok := err.Err.(*AppError); ok {
				appErr = e
			} else {
				// Wrap unknown errors
				appErr = ErrInternalError("An unexpected error occurred", err.Err)
			}

			// Log the error
			logger.Error("Request error",
				zap.String("code", string(appErr.Code)),
				zap.String("sub_code", string(appErr.SubCode)),
				zap.String("message", appErr.Message),
				zap.String("detail", appErr.Detail),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Error(appErr.Err),
			)

			// Return error response
			c.JSON(appErr.HTTPCode, ErrorResponse{
				Error: ErrorInfo{
					Code:    string(appErr.Code),
					SubCode: string(appErr.SubCode),
					Message: appErr.Message,
					Detail:  appErr.Detail,
				},
			})
		}
	}
}

// HandleError is a helper function to handle errors in handlers
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// If it's already an AppError, add it to context
	if appErr, ok := err.(*AppError); ok {
		c.Error(appErr)
		c.Abort()
		return
	}

	// Wrap unknown errors
	appErr := ErrInternalError("An unexpected error occurred", err)
	c.Error(appErr)
	c.Abort()
}

// AbortWithError aborts the request with an error
func AbortWithError(c *gin.Context, err *AppError) {
	c.Error(err)
	c.Abort()
	c.JSON(err.HTTPCode, ErrorResponse{
		Error: ErrorInfo{
			Code:    string(err.Code),
			SubCode: string(err.SubCode),
			Message: err.Message,
			Detail:  err.Detail,
		},
	})
}

