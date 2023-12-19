package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CtxLogger logs the request context
func CtxLogger(ctx *gin.Context) {
	logger := GetLogger(ctx)
	address := ctx.ClientIP()
	headerCopy := make(map[string][]string)
	for k, v := range ctx.Request.Header {
		headerCopy[k] = make([]string, len(v))
		copy(headerCopy[k], v)
	}
	headersStr := fmt.Sprintf("%v", headerCopy)

	// execute handlers.
	ctx.Next()

	// get status.
	statusCode := ctx.Writer.Status()
	logger = logger.With(
		zap.String("address", address),
		zap.Int("status_code", statusCode),
	)
	if statusCode >= 200 && statusCode < 300 {
		logger.Info("req_done")
	} else {
		logger.Warn("req_done")
		logger.With(zap.String("headers", headersStr)).Debug("req_debug_info")
	}
}
