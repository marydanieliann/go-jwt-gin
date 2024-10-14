package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go-jwt/initializers"
	"go-jwt/models"
	"net/http"
	"os"
	"time"
)

func RequireAuth(c *gin.Context) {
	fmt.Println("In middleware")
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx := c.Request.Context()
		exists, err := initializers.RedisClient.HExists(ctx, "user_tokens", tokenString).Result()
		if err != nil {
			fmt.Println("Error checking token existence in Redis:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if exists == false {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var user models.User
		if user.ID < 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
