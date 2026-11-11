package auth

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Account struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	PasswordHash  string `json:"-" db:"password_hash"`
	EmailVerified bool   `json:"-" db:"email_verified"`
}
