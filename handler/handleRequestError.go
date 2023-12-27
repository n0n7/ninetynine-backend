package handler

import (
	"encoding/json"
	"net/http"
)

func requestErrorHandler(w http.ResponseWriter, errMessage string, statusCode int) {
	responseData := map[string]interface{}{
		"error": errMessage,
	}

	responseJSON, _ := json.Marshal(responseData)
	w.WriteHeader(statusCode)
	w.Write(responseJSON)
}
