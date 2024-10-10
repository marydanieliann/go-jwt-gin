package main

import (
	"github.com/gin-gonic/gin"
	"go-jwt/controllers"
	"go-jwt/initializers"
	"go-jwt/middleware"
)

// loads before main
func init() {
	initializers.LoadEnvVariables()
	//the below 2 lines are used for connection to database
	initializers.ConnectToDb()
	initializers.SyncDatabase()
	initializers.ConnectRedis()
}

func main() {
	r := gin.Default()
	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)
	r.GET("/validate", middleware.RequireAuth, controllers.Validate)
	r.POST("/movie", controllers.CreateMovie)
	r.GET("/movie/:id", controllers.GetMovieById)
	r.GET("/movies", controllers.GetAllMovies)
	r.Run()
}
