package main

import (
	"fmt"
	"gin-app/param"
	"gin-app/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/get/sandbox", func(c *gin.Context) {
		reqSandbox := param.ReqSandbox{}
		if err := c.ShouldBind(&reqSandbox); err == nil {
			fmt.Printf("KeyCl: %v\n", reqSandbox.KeyCl)
			list, err2 := service.GetSandboxList(reqSandbox)
			if err2 != nil {
				fmt.Printf("エラー: %v\n", err2)
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "error",
				})
			}
			c.JSON(http.StatusOK, list)
		} else {
			fmt.Printf("エラー: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "error",
			})
		}
	})

	r.Run() // 0.0.0.0:8080 でサーバーを立てます。
}
