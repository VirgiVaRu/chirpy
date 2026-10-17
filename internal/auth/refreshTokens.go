package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() string {
	randomData := make([]byte, 32)
	rand.Read(randomData)
	return hex.EncodeToString(randomData)
}
