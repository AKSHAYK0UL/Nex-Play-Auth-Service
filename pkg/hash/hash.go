package hash

import "golang.org/x/crypto/bcrypt"

const cost = 12

// Hash password
func Password(plain string) (string, error) {

	b, err := bcrypt.GenerateFromPassword([]byte(plain), cost)

	if err != nil {

		return "", err
	}

	return string(b), nil
}

//Check Hash
func CheckPassword(hashed, plain string) bool {

	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
