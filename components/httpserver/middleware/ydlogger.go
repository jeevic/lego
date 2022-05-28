package middleware

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

func YdLoggerMiddleWare(output io.Writer, hostIp string) gin.HandlerFunc {
	logCfg := gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			//兼容新一点格式 时间|host|
			format := "%s|info|%s|0|requestId=%s, client-ip=%s, method=%s, path=%s, proto=%s, statusCode=%d, bodySize=%d, latency=%s, user-agent=%s, http-referer=%s, error-message=%s \n"
			return fmt.Sprintf(format,
				param.TimeStamp.Format("2006-01-02 15:04:05.000"),
				hostIp,
				param.Request.Header.Get("X-Request-Id"),
				param.ClientIP,
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.BodySize,
				param.Latency,
				param.Request.UserAgent(),
				param.Request.Referer(),
				param.ErrorMessage,
			)
		},
		Output: output,
	}
	return gin.LoggerWithConfig(logCfg)
}
