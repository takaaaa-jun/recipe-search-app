package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	r := GinRouter()
	r.Run(":8000") // 8000番ポートで起動
}

func GinRouter() *gin.Engine {
	r := gin.Default()
	
	// ヘルスチェック用
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello!",
		})
	})
	
	return r
}