package generateotp

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// generate new OTP
func New(digits int) (string, error) {

	mod := uint32(1)

	for i := 0; i < digits; i++ {
		mod *= 10
	}

	var b [4]byte

	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	n := binary.BigEndian.Uint32(b[:]) % mod

	return fmt.Sprintf("%0*d", digits, n), nil
}
