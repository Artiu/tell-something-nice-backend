package main

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tell-something-nice-backend/controllers"
	"github.com/tell-something-nice-backend/models"
)

func convertAuthorizationArray(c *gin.Context) {
	var tokens []string
	tokensStringArr := c.Request.Header["Authorization"]
	tokenString := strings.Join(tokensStringArr, "")
	json.Unmarshal([]byte(tokenString), &tokens)

	c.Set("tokens", tokens)
}

func main() {
	godotenv.Load(".env")

	r := gin.Default()

	models.ConnectToDB()
	err := models.DB.Ping()
	if err != nil {
		panic(err)
	}
	posts := r.Group("/posts", convertAuthorizationArray)
	posts.GET("/", controllers.GetPosts)
	posts.POST("/", controllers.AddPost)
	posts.PATCH("/:id", controllers.EditPost)
	posts.DELETE("/:id", controllers.RemovePost)

	r.Run("localhost:8080")
}
