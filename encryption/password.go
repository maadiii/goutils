package encryption

import "golang.org/x/crypto/bcrypt"

type password struct{}

func NewPassword() *password {
	return new(password)
}

func (p *password) Generate(plain string) (hash string, err error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	hash = string(bytes)

	return
}

func (p *password) Compare(hash string, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))

	return err == nil
}
