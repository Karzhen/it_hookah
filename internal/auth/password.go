package auth

import "golang.org/x/crypto/bcrypt"

type PasswordManager interface {
	HashPassword(password string) (string, error)
	ComparePassword(hash string, password string) error
}

type BcryptPasswordManager struct {
	cost int
}

func NewBcryptPasswordManager(cost int) *BcryptPasswordManager {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	return &BcryptPasswordManager{cost: cost}
}

func (m *BcryptPasswordManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), m.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (m *BcryptPasswordManager) ComparePassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
