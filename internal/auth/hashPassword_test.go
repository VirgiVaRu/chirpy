package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	cases := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "password matches hash",
			password: "correctPassword",
			expected: true,
		},
		{
			name:     "password does not match hash",
			password: "fakePassword",
			expected: false,
		},
	}

	hash, err := HashPassword("correctPassword")
	if err != nil {
		t.Fatalf("unable to hash password: %s", err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := CheckPasswordHash(c.password, hash)
			if err != nil {
				t.Fatalf("error comparing password: %s", err)
			}
			if actual != c.expected {
				t.Errorf("got: %v\nwant: %v\n", actual, c.expected)
			}
		})
	}
}
