package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {

	userID := uuid.New()
	cases := []struct {
		name        string
		tokenSecret string
		timePassed  time.Duration
		expected    uuid.UUID
		expectedErr bool
	}{
		{
			name:        "matching token",
			tokenSecret: "thisIsRight",
			timePassed:  0 * time.Millisecond,
			expected:    userID,
			expectedErr: false,
		},
		{
			name:        "unmatched token",
			tokenSecret: "thisIsWrong",
			timePassed:  0 * time.Millisecond,
			expected:    uuid.Nil,
			expectedErr: true,
		},
		{
			name:        "expired token",
			tokenSecret: "thisIsRight",
			timePassed:  1000 * time.Millisecond,
			expected:    uuid.Nil,
			expectedErr: true,
		},
	}

	expiresIn := time.Duration(1000 * time.Millisecond)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tokenString, err := MakeJWT(userID, "thisIsRight", expiresIn)
			if err != nil {
				t.Fatalf("error creating the JWT: %s", err)
			}
			time.Sleep(c.timePassed)
			actual, err := ValidateJWT(tokenString, c.tokenSecret)
			if actual != c.expected || (err != nil) != c.expectedErr {
				t.Errorf("got: %v\nwant: %v\nerror found: %v\n", actual, c.expected, err)
			}
		})
	}
}

func headerBuilder(key, value string) http.Header {
	h := http.Header{}
	h.Set(key, value)
	return h
}

func TestGetBearerToken(t *testing.T) {

	cases := []struct {
		name           string
		header         http.Header
		expected       string
		expectedErrors bool
	}{
		{
			name:           "everything is correct",
			header:         headerBuilder("Authorization", "Bearer token_string"),
			expected:       "token_string",
			expectedErrors: false,
		},
		{
			name:           "no authentication found",
			header:         headerBuilder("Content-Type", "application/json"),
			expected:       "",
			expectedErrors: true,
		},
		{
			name:           "wrong format",
			header:         headerBuilder("Authorization", "Something token_string"),
			expected:       "",
			expectedErrors: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := GetBearerToken(c.header)
			if (err != nil) != c.expectedErrors {
				t.Errorf("found %v when expected errors were: %v\n", err, c.expectedErrors)
			}
			if actual != c.expected {
				t.Errorf("got: %v\nexpected: %v\n", actual, c.expected)
			}
		})
	}
}
