package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIdMiddleware(requestIdName string) gin.HandlerFunc {
	if len(requestIdName) == 0 {
		requestIdName = "X-Request-Id"
	}
	return func(c *gin.Context) {
		//判断获取uuid
		u := c.GetHeader(requestIdName)
		if len(u) < 1 {
			u = uuid.New().String()
		}
		//设置request id
		c.Request.Header.Set(requestIdName, u)
		c.Writer.Header().Set(requestIdName, u)

		c.Next()
	}
}
