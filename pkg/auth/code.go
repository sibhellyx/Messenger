package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateLoginCode generate code for verify login
func GenerateLoginCode() string {
	// generate 6-code
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n)
}
