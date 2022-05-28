package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// cors middleware
func CorsMiddleWare() gin.HandlerFunc {
	//TODO:: customize your own CORS
	//https://github.com/gin-contrib/cors
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, //https://foo.com
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, //enable cookie
		AllowOriginFunc: func(origin string) bool {
			return true
			//return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour, //cache options result decrease request lag
	})
}
