package main

import (
	"backend/api/context"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine) {

	pprof.Register(r)

	context.GetPageRefApi().RegisterRoutes(r)
	context.GetWebsiteApi().RegisterRoutes(r)
	context.GetScheduleApi().RegisterRoutes(r)

	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())
}
