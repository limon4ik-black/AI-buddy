package auth

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Account struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
}

type PendingRegistration struct {
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}
