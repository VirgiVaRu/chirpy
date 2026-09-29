package main

import (
	"strings"
)

func removeProfanity(msg string) string {
	profanity := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(msg, " ")
	for i, word := range words {
		for _, badWord := range profanity {
			if strings.ToLower(word) == badWord {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}
