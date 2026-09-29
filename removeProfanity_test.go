package main

import (
	"testing"
)

func TestRemoveProfanity(t *testing.T) {
	cases := []struct {
		name     string
		msg      string
		expected string
	}{
		{
			name:     "nothing to change",
			msg:      "There is nothing to change here.",
			expected: "There is nothing to change here.",
		},
		{
			name:     "this should be censored",
			msg:      "This is a kerfuffle opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := removeProfanity(c.msg)
			if actual != c.expected {
				t.Errorf("got: %s\nwant: %s\n", actual, c.expected)
			}
		})
	}
}
