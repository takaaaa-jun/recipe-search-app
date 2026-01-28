package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"backend/db"
	"backend/handlers"
)

func main() {
	database := db.Init()
	defer database.Close()

	personalRecipeHandler := handlers.PersonalRecipeHandler{DB: database}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-CSRF-Token, Authorization")
		c.Next()
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
	})

	r.GET("/personal-recipes", personalRecipeHandler.PersonalRecipes)

	r.Run(":8000")
}