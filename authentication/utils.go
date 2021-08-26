package authentication

import "golang.org/x/crypto/bcrypt"

var DefaultCost = bcrypt.DefaultCost

func Hash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, DefaultCost)
}

func Compare(hash, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hash, password) == nil
}
