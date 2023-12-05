package handlers

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	// "io"
	"log"
	// "mime/multipart"
	"net/http"
	// "time"

	// "cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"

	"google.golang.org/api/option"
)

func UploadFile(c *gin.Context) {
	ctx := context.Background()

	// Set up Google Cloud Storage client
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialFile))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Cloud Storage client"})
		return
	}
	defer client.Close()

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	log.Printf("Authenticated user: %v", user)

	// Parse the form data to get the file part
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Get the file from the request
	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	file := files[0]

	// Create a unique filename for the uploaded file
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)

	// Set up the destination object in the Cloud Storage bucket
	object := client.Bucket(bucketName).Object(filename)

	// Create a writer to the object
	wc := object.NewWriter(ctx)

	// Open the file on the server
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file on the server"})
		return
	}
	defer src.Close()

	// Copy the file content to the Cloud Storage object
	if _, err := io.Copy(wc, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy file content to Cloud Storage"})
		return
	}

	// Close the writer to finalize the upload
	if err := wc.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close Cloud Storage writer"})
		return
	}

	// // Optionally, you can set the ACL for the uploaded object
	if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set ACL for Cloud Storage object"})
		return
	}

	// Return the URL of the uploaded file
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filename)
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "url": url})
}

// Helper function to get the file extension
func getFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// func UploadFile(c *gin.Context) {
// 	// Check if the user is logged in
// 	userEmail, loggedIn := getLoggedInUser(c)
// 	if !loggedIn {
// 		log.Printf("User not logged in. Request: %v", c.Request)
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
// 		return
// 	}

// 	// Get the uploaded file from the request
// 	file, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from the request"})
// 		return
// 	}

// 	// Set up Google Cloud Storage client
// 	ctx := context.Background()
// 	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialFile))
// 	if err != nil {
// 		log.Fatalf("Failed to create client: %v", err)
// 	}
// 	defer client.Close()

// 	// Upload the file to the Google Cloud Storage bucket
// 	objectName := generateObjectName(userEmail, file.Filename)
// 	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
// 	defer wc.Close()

// 	if _, err := wc.Write([]byte("")); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to storage bucket"})
// 		return
// 	}

// 	if err := wc.Close(); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close storage bucket writer"})
// 		return
// 	}

// 	// Update the access list in Firestore
// 	if err := updateAccessList(ctx, userEmail, objectName); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update access list in Firestore"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{"message": "File uploaded successfully"})
// }

// // generateObjectName generates a unique object name based on user email and filename
// func generateObjectName(userEmail, filename string) string {
// 	return fmt.Sprintf("%s/%s", userEmail, filename)
// }

// // updateAccessList adds the object to the access list of the user in Firestore
// func updateAccessList(ctx context.Context, userEmail, objectName string) error {
// 	firestoreClient, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialFile))
// 	if err != nil {
// 		return fmt.Errorf("failed to create Firestore client: %v", err)
// 	}
// 	defer firestoreClient.Close()

// 	// Get the current access list
// 	accessList, err := getAccessList(ctx, firestoreClient, userEmail)
// 	if err != nil {
// 		return fmt.Errorf("failed to get access list: %v", err)
// 	}

// 	// Add the new object to the access list
// 	accessList = append(accessList, objectName)

// 	// Update the access list in Firestore
// 	_, err = firestoreClient.Collection("users").Doc(userEmail).Set(ctx, map[string]interface{}{
// 		"AccessList": accessList,
// 	}, firestore.MergeAll)
// 	if err != nil {
// 		return fmt.Errorf("failed to update access list in Firestore: %v", err)
// 	}

// 	return nil
// }

// // getAccessList retrieves the current access list of the user from Firestore
// func getAccessList(ctx context.Context, client *firestore.Client, userEmail string) ([]string, error) {
// 	doc, err := client.Collection("users").Doc(userEmail).Get(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get user document from Firestore: %v", err)
// 	}

// 	var user map[string]interface{}
// 	if err := doc.DataTo(&user); err != nil {
// 		return nil, fmt.Errorf("failed to parse user data: %v", err)
// 	}

// 	accessList, ok := user["AccessList"].([]interface{})
// 	if !ok {
// 		return nil, fmt.Errorf("AccessList field not found or has invalid type")
// 	}

// 	var accessListStrings []string
// 	for _, item := range accessList {
// 		if s, ok := item.(string); ok {
// 			accessListStrings = append(accessListStrings, s)
// 		} else {
// 			return nil, fmt.Errorf("invalid item in AccessList")
// 		}
// 	}

// 	return accessListStrings, nil
// }

// // getLoggedInUser retrieves the logged-in user's email from the context
// func getLoggedInUser(c *gin.Context) (string, bool) {
// 	userEmail, exists := c.Get("userEmail")
// 	if !exists {
// 		return "", false
// 	}
// 	return userEmail.(string), true
// }
