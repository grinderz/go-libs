package librandom

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Hex returns a hex-encoded string of bytesLen random bytes,
// so the resulting string is 2*bytesLen characters long.
func Hex(bytesLen int) (string, error) {
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}

	return hex.EncodeToString(b), nil
}
