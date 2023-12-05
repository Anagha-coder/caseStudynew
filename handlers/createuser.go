package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"casestudy/models"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	projectID      = "casestudycs"
	credentialFile = "./casestudycs.json"
	bucketName     = "maxstoragebucketformanyusersoncloud"
)

func CreateUser(c *gin.Context) {
	ctx := context.Background()

	var newUser models.User

	// Set up Firestore client
	firestoreClient, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialFile))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Set up Google Cloud Storage client
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialFile))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Bind JSON request body to the User struct
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the email is unique
	if exists, err := isEmailUnique(ctx, firestoreClient, newUser.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email uniqueness"})
		return
	} else if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email address is already in use"})
		return
	}

	// Check if the username is unique
	if exists, err := isUsernameUnique(ctx, firestoreClient, newUser.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username uniqueness"})
		return
	} else if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is already in use"})
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Save user details in Firestore
	userID, err := addUserToFirestore(ctx, firestoreClient, newUser, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "userID": userID})

}

func isEmailUnique(ctx context.Context, client *firestore.Client, email string) (bool, error) {
	iter := client.Collection("users").Where("Email", "==", email).Limit(1).Documents(ctx)
	defer iter.Stop()

	if _, err := iter.Next(); err != nil {
		if err == iterator.Done {
			// No matching documents, email is unique
			return true, nil
		}
		// An error occurred while fetching the next document
		return false, fmt.Errorf("error checking email uniqueness: %v", err)
	} else {
		// Found a matching document, email is not unique
		return false, nil
	}
}

func isUsernameUnique(ctx context.Context, client *firestore.Client, username string) (bool, error) {
	iter := client.Collection("users").Where("Username", "==", username).Limit(1).Documents(ctx)
	defer iter.Stop()

	if _, err := iter.Next(); err != nil {
		if err == iterator.Done {
			// No matching documents, username is unique
			return true, nil
		}
		// An error occurred while fetching the next document
		return false, fmt.Errorf("error checking username uniqueness: %v", err)
	} else {
		// Found a matching document, username is not unique
		return false, nil
	}
}

func addUserToFirestore(ctx context.Context, client *firestore.Client, user models.User, hashedPassword string) (string, error) {
	// Add a new document to the "users" collection in Firestore

	// Use the email as the document ID
	docRef := client.Collection("users").Doc(user.Email)

	// Set the data for the user
	_, err := docRef.Set(ctx, map[string]interface{}{
		"Name":      user.Name,
		"Username":  user.Username,
		"Email":     user.Email,
		"Password":  hashedPassword,
		"CreatedAt": time.Now(),
		"Access":    user.AccessList,
	})
	if err != nil {
		return "", fmt.Errorf("failed to add user to Firestore: %v", err)
	}

	// Return the document ID as the user ID
	return docRef.ID, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
