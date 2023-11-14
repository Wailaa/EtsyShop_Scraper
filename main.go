package main

import (
	initializer "EtsyScraper/init"
	"net/http"

	"github.com/gin-gonic/gin"

	"fmt"
	"log"
)

var server *gin.Engine

func init() {
	config, err := initializer.LoadProjConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	initializer.DataBaseConnect(&config)
}
func main() {

	config, err := initializer.LoadProjConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)

	}

	server = gin.Default()
	fmt.Println("Migration is completed")

	router := server.Group("/auth")
	router.GET("/test", func(ctx *gin.Context) {
		message := "Welcome to EtsyScraper a"
		ctx.JSON(http.StatusOK, gin.H{"HTTPstatus": http.StatusOK, "message": message})
	})

	log.Fatal(server.Run(":" + config.ServerPort))
}
