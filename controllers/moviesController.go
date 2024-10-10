package controllers

import (
	"github.com/gin-gonic/gin"
	"go-jwt/models"
	"go-jwt/services"
	"net/http"
)

func CreateMovie(c *gin.Context) {
	var body struct {
		Title    string `json:"title"`
		Director string `json:"director"`
		UserID   uint   `json:"user_id"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	movie := models.Movie{Title: body.Title, Director: body.Director, UserID: body.UserID}

	//normal DB connection
	/*
		result := initializers.DB.Create(&movie)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create new movie"})
			return
		}
	*/

	//Redis part
	if err := services.CacheMovie(c, movie); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cache movie"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"movie": movie})
}

func GetAllMovies(c *gin.Context) {
	movies, err := services.GetAllMoviesFromCache(c)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"movies": movies})
		return
	}

	var dbMovies []models.Movie
	//normal DB connection
	/*
		result := initializers.DB.Find(&dbMovies)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get all movies"})
			return
		}
	*/

	//Redis part
	if err := services.CacheAllMovies(c, dbMovies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cache movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": dbMovies})
}

func GetMovieById(c *gin.Context) {
	movieID := c.Param("id")

	movie, err := services.GetMovieFromCache(c, movieID)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"movie": movie})
		return
	}

	var movieFromDB models.Movie
	//normal DB connection
	/*
		result := initializers.DB.First(&movieFromDB, movieID)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get movie with id " + movieID})
			return
		}
	*/

	//Redis part
	if err := services.CacheMovie(c, movieFromDB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cache movie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": movieFromDB})
}
