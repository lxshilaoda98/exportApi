package main

import (
	. "exportApi/servives"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"net/http"
	"time"
)

func RateLimitMiddleware(fillInterval time.Duration, cap, quantum int64) gin.HandlerFunc {
	bucket := ratelimit.NewBucketWithQuantum(fillInterval, cap, quantum)
	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			c.String(http.StatusForbidden, "rate limit...")
			c.Abort()
			return
		}
		c.Next()
	}
}

func InitRouter() *gin.Engine {
	router := gin.Default()

	//跨域
	//router.Use(Cors())

	//router.GET("/dispatchInfoStart", DispatchInfoStart)
	exportApi := router.Group("export")
	{
		exportApi.Use(RateLimitMiddleware(time.Second, 10, 1)) //初始10，每秒放出10
		exportApi.POST("/icrm", ExportIcrm)
	}

	return router
}
