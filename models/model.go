package models

type User struct {
	Name       string   `firestore:"name"`
	Username   string   `firestore:"username"`
	Email      string   `firestore:"email"`
	Password   string   `firestore:"password"`
	BucketName string   `firestore:"bucketName"`
	AccessList []string `firestore:"access"`
}

// for user login requests
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
