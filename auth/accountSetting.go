package auth

import (
	"context"
	Firebase "ninetynine/firebase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountData struct {
	UserId   string `json:"userId"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func AccountSetting(data AccountData) (bool, map[string]interface{}, error) {
	docRef := Firebase.FirestoreClient.Collection("users").Doc(data.UserId)
	docSnap, err := docRef.Get(context.Background())

	if status.Code(err) == codes.NotFound {
		return false, nil, nil
	}

	if err != nil {
		return false, nil, err
	}

	userData := docSnap.Data()

	userData["email"] = data.Email
	userData["username"] = data.Username

	_, err = docRef.Set(context.Background(), userData)
	if err != nil {
		return false, nil, err
	}

	return true, userData, nil

}

func resetPassword() {
	// TODO: implement
}
