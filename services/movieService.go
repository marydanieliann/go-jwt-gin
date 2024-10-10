package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go-jwt/initializers"
	"go-jwt/models"
	"strconv"
	"time"
)

func CacheMovie(ctx context.Context, movie models.Movie) error {
	movieData, err := json.Marshal(movie)
	if err != nil {
		return fmt.Errorf("failed to serialize movie: %v", err)
	}

	key := "movie:" + strconv.Itoa(int(movie.ID))
	return initializers.RedisClient.Set(ctx, key, movieData, time.Hour*24).Err()
}

func GetMovieFromCache(ctx context.Context, movieID string) (*models.Movie, error) {
	movieData, err := initializers.RedisClient.Get(ctx, "movie:"+movieID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("movie not found in cache")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get movie from cache: %v", err)
	}

	var movie models.Movie
	err = json.Unmarshal([]byte(movieData), &movie)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize movie data: %v", err)
	}
	return &movie, nil
}

func CacheAllMovies(ctx context.Context, movies []models.Movie) error {
	movieData, err := json.Marshal(movies)
	if err != nil {
		return fmt.Errorf("failed to serialize movies: %v", err)
	}

	return initializers.RedisClient.Set(ctx, "movies:all", movieData, time.Hour*24).Err()
}

func GetAllMoviesFromCache(ctx context.Context) ([]models.Movie, error) {
	movieData, err := initializers.RedisClient.Get(ctx, "movies:all").Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("movies not found in cache")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get movies from cache: %v", err)
	}

	var movies []models.Movie
	err = json.Unmarshal([]byte(movieData), &movies)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize movie data: %v", err)
	}
	return movies, nil
}
