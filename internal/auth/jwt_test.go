package auth

import (
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
