package hash

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

type Hasher struct {
	salt string
}

func NewHasher(salt string) *Hasher {
	return &Hasher{salt: salt}
}

func (h *Hasher) Hash(password string) (string, error) {
	hash := sha512.New()

	if _, err := hash.Write([]byte(password)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum([]byte(h.salt))), nil
}

func (h *Hasher) HashRefreshToken(refreshToken string) string {
	return base64.StdEncoding.EncodeToString([]byte(refreshToken))
}
