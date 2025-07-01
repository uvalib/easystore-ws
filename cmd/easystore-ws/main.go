package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func main() {

	log.Printf("===> %s service staring up (version: %s) <===", os.Args[0], Version())

	// Get config params and use them to init service context. Any issues are fatal
	cfg := LoadConfiguration()
	svc := NewService(cfg)

	gin.SetMode(gin.ReleaseMode)
	//gin.SetMode(gin.DebugMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	//corsCfg := cors.DefaultConfig()
	//corsCfg.AllowAllOrigins = true
	//corsCfg.AllowCredentials = true
	//corsCfg.AddAllowHeaders("Authorization")
	//router.Use(cors.New(corsCfg))

	router.GET("/", svc.GetVersion)
	router.GET("/favicon.ico", svc.IgnoreFavicon)
	router.GET("/version", svc.GetVersion)
	router.GET("/healthcheck", svc.HealthCheck)

	// get a single object
	router.GET("/:ns/:id", svc.GetObject)
	// get many objects
	router.PUT("/:ns", svc.GetObjects)
	// object search
	router.PUT("/:ns/search", svc.SearchObjects)
	// create a new object
	router.POST("/:ns", svc.CreateObject)
	// update an existing object
	router.PUT("/:ns/:id", svc.UpdateObject)
	// rename a blob from an existing object
	router.POST("/:ns/:id", svc.RenameBlob)
	// delete an existing object
	router.DELETE("/:ns/:id", svc.DeleteObject)

	portStr := fmt.Sprintf(":%d", cfg.Port)
	log.Fatal(router.Run(portStr))
}

//
// end of file
//
