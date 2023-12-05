package handlers

import (
	"casestudy/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func Login(c *gin.Context) {
	ctx := context.Background()
	var loginReq models.LoginRequest

	// Set up Firestore client
	firestoreClient, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialFile))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := getUserByEmail(loginReq.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user. Check provided email id again"})
		return
	}

	// Check if the provided password matches the stored hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate and send a JWT token
	token, err := generateToken(user, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Login successful"})
}

func getUserByEmail(email string) (*models.User, error) {
	ctx := context.Background()

	// Set up Firestore client
	firestoreClient, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialFile))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	iter := firestoreClient.Collection("users").Where("Email", "==", email).Limit(1).Documents(ctx)
	defer iter.Stop()

	for {
		snapshot, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				// email not found
				return nil, fmt.Errorf("user with email %s not found", email)
			}
			// error while fetching the next document
			return nil, fmt.Errorf("error fetching user: %v", err)
		}

		// Found a matching document, parse it into a User struct
		var user models.User
		if err := snapshot.DataTo(&user); err != nil {
			return nil, fmt.Errorf("error parsing user data: %v", err)
		}

		return &user, nil
	}

}

// func generateToken(user *models.User, c *gin.Context) (string, error) {
// 	// Set expiration time for the token (e.g., 24 hours)
// 	expirationTime := time.Now().Add(24 * time.Hour)

// 	// Create a simple JWT with username and expiration time
// 	claims := jwt.MapClaims{
// 		"username": user.Username,
// 		"email":    user.Email,
// 		"exp":      expirationTime.Unix(),
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	signedToken, err := token.SignedString([]byte("./secretKey.txt"))
// 	if err != nil {
// 		return "", fmt.Errorf("error signing token: %v", err)
// 	}

// 	// Set the user's email in the Gin context
// 	c.Set("userEmail", user.Email)

// 	fmt.Println("User's email set in the context:", user.Email)

// 	fmt.Println("Generated Token:", signedToken)

// 	return signedToken, nil

// }

func generateToken(user *models.User, c *gin.Context) (string, error) {
	// Set expiration time for the token (e.g., 24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create a simple JWT with username and expiration time
	claims := jwt.MapClaims{
		"username": user.Username,
		"email":    user.Email,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("./secretKey.txt"))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	// Set the user's email in the Gin context
	c.Set("userEmail", user.Email)
	fmt.Println("User's email set in the context:", user.Email)

	fmt.Println("Generated Token:", signedToken)

	return signedToken, nil
}
