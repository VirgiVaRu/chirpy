package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authorizationString := headers.Get("Authorization")
	if authorizationString == "" {
		return authorizationString, errors.New("Couldn't find an authorization string")
	}
	if strings.HasPrefix(authorizationString, "ApiKey ") == false {
		return "", errors.New("Authorization has the wrong format")
	}
	apiKey := strings.TrimPrefix(authorizationString, "ApiKey ")
	return apiKey, nil
}
