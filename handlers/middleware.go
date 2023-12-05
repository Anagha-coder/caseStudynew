package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const secretKey = "./secretkey.txt"

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractTokenFromHeader(c.Request.Header.Get("Authorization"))
		fmt.Println("Token:", tokenString)

		token, err := verifyToken(tokenString)
		if err != nil {
			fmt.Println("Token verification error:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Attach the user information to the context
		c.Set("user", token.Claims.(jwt.MapClaims)["email"])

		// Continue with the next handler
		c.Next()
	}
}

func extractTokenFromHeader(header string) string {
	// Check if the Authorization header is present
	if header == "" {
		return ""
	}

	// Extract the token from the header (assuming it is a Bearer token)
	parts := strings.Split(header, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	fmt.Println("Token received:", tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		fmt.Println("Token parsing error:", err) // Add this line for debugging
		return nil, err
	}

	if !token.Valid {
		fmt.Println("Token is not valid") // Add this line for debugging
		return nil, jwt.ErrSignatureInvalid
	}
	return token, nil
}
