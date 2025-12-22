package firebase

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var App *firebase.App

// Init Firebase App (call once)
func InitFirebase() {
	path := os.Getenv("SERVICE_ACCOUNT_JSON")
	if path == "" {
		log.Fatal("SERVICE_ACCOUNT_JSON not set")
	}

	opt := option.WithCredentialsFile(path)

	var err error
	App, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Firebase init failed: %v", err)
	}

	log.Println("✅ Firebase initialized")
}

// Get Firebase Auth client
// func GetAuthClient() *auth.Client {

func GetAuthClient() *auth.Client {
	if App == nil {
		log.Fatal("❌ Firebase not initialized. Call InitFirebase() first")
	}

	authClient, err := App.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	// // Example: Create a user (for testing)
	// params := (&auth.UserToCreate{}).
	// 	Email("user@gmail.com").
	// 	Password("56489649").
	// 	DisplayName("serd")

	//u, err := authClient.CreateUser(context.Background(), params)
	if err != nil {
		log.Fatalf("error creating user: %v\n", err)
	}
	//fmt.Printf("Successfully created user: UID=%s, DisplayName=%s\n", u.UID, u.DisplayName)

	return authClient
}
