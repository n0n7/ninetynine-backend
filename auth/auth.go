package firebase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"

	Firebase "ninetynine/firebase"

	"google.golang.org/api/iterator"
)

func CheckUniqueEmail(email string) (bool, error) {
	query := Firebase.FirestoreClient.Collection("users").Where("email", "==", email).Limit(1)
	_, err := query.Documents(context.Background()).Next()
	if err == iterator.Done {
		return true, nil
	}
	if err != nil && err != iterator.Done {
		return false, err
	}
	return false, nil
}

func CheckUniqueUsername(username string) (bool, error) {
	query := Firebase.FirestoreClient.Collection("users").Where("username", "==", username).Limit(1)
	_, err := query.Documents(context.Background()).Next()
	if err == iterator.Done {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	return false, nil
}

func CreateUser(userData map[string]interface{}) (map[string]interface{}, error) {
	// hash password
	userData["password"] = hashedPassword(userData["password"].(string))

	docRef, _, err := Firebase.FirestoreClient.Collection("users").Add(context.Background(), userData)
	if err != nil {
		return nil, err
	}

	userData["userId"] = docRef.ID
	userData["userId"] = docRef.ID
	Firebase.FirestoreClient.Collection("users").Doc(docRef.ID).Set(context.Background(), userData)

	delete(userData, "password")
	return userData, nil
}

func Login(email string, password string) (bool, map[string]interface{}, error) {
	query := Firebase.FirestoreClient.Collection("users").Where("email", "==", email).Limit(1)
	docSnap, err := query.Documents(context.Background()).Next()
	if err == iterator.Done {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	userData := docSnap.Data()
	if userData["password"] != hashedPassword(password) {
		return false, nil, nil
	}

	delete(userData, "password")
	return true, userData, nil
}

func hashedPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed)
}

func IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	return match
}
