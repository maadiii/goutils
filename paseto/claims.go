package paseto

import "time"

type Claims struct {
	Audience    string
	Expiration  time.Time
	Jti         string
	Subject     string
	PhoneNumber string
	TotpSecret  string
	StateToken  string
	UserID      string
	Role        string
}

func (t *Claims) SetAudience(value string) *Claims {
	t.Audience = value

	return t
}

func (t *Claims) SetExpiration(value time.Time) *Claims {
	t.Expiration = value

	return t
}

func (t *Claims) SetJti(value string) *Claims {
	t.Jti = value

	return t
}

func (t *Claims) SetSubject(value string) *Claims {
	t.Subject = value

	return t
}

func (t *Claims) SetPhoneNumber(value string) *Claims {
	t.PhoneNumber = value

	return t
}

func (t *Claims) SetTotpSecret(value string) *Claims {
	t.TotpSecret = value

	return t
}

func (t *Claims) SetStateToken(value string) *Claims {
	t.StateToken = value

	return t
}

func (t *Claims) SetUserID(value string) *Claims {
	t.UserID = value

	return t
}

func (t *Claims) SetRole(value string) *Claims {
	t.Role = value

	return t
}
