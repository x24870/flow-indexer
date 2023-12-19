package middlewares

import (
	"context"
	"flow-indexer/pkg/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetLogger gets logger from gin context and returns log.Null() as the
// default value.
func GetLogger(c *gin.Context) *zap.Logger {
	logger := log.GetFromCtx(c.Request.Context())
	if logger != log.Nop {
		return logger
	}

	var ctx context.Context
	ctx, logger = log.Context(c.Request.Context())
	logger = logger.With(
		zap.String("gin_method", c.Request.Method),
		zap.String("gin_url", c.Request.URL.String()),
	)
	ctx = context.WithValue(ctx, log.CtxKey(), logger)
	c.Request = c.Request.WithContext(ctx)
	return logger
}
