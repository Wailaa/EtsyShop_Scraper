package controllers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func HandleResponse(ctx *gin.Context, err error, status int, message string, data interface{}) {
	var response gin.H
	if err != nil {
		log.Println(err)
	}

	if data != nil {
		ctx.JSON(status, data)
		return
	}

	response = gin.H{"status": "fail"}
	if status == 200 {
		response = gin.H{"status": "success"}
	}

	if message != "" {
		response["message"] = message
	}

	ctx.JSON(status, response)

	if status != 200 {
		ctx.Abort()
	}

}
