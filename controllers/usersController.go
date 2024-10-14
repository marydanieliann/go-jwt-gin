package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"go-jwt/initializers"
	"go-jwt/models"
	"go-jwt/services"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

func Signup(c *gin.Context) {
	var body struct {
		Email    string
		Password string
	}
	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.MinCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{Email: body.Email, Password: string(hash)}

	//normal DB connection
	/*
		result := initializers.DB.Create(&user)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user"})
			return
		}
	*/

	// Redis part
	ctx := context.Background()
	if err := services.UserToCache(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cache user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string
		Password string
	}
	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	ctx := context.Background()
	var user *models.User
	user, err := services.GetUserFromCache(ctx, body.Email)
	if err != nil {
		//normal DB connection
		/*
			if err := initializers.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Invalid email or password"})
				return
			}
		*/
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	err = initializers.RedisClient.HSet(ctx, "user_tokens", tokenString, user.ID).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{})
}

func Validate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func Logout(c *gin.Context) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active session"})
		return
	}

	fmt.Println("Token from cookie:", tokenString)
	ctx := c.Request.Context()
	_, err = initializers.RedisClient.HGet(ctx, "user_tokens", tokenString).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			fmt.Println("Token not found in user_tokens for:", tokenString)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token not found"})
			return
		}
		fmt.Println("Error getting user ID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user ID"})
		return
	}

	err = initializers.RedisClient.HDel(ctx, "user_tokens", tokenString).Err()
	if err != nil {
		fmt.Println("Error deleting token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete token"})
		return
	} else {
		fmt.Println("Successfully deleted token:", tokenString)
	}

	c.SetCookie("Authorization", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
