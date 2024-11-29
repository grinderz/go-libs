package librandom

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func String(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}

	return hex.EncodeToString(b), nil
}
