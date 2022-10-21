package model

type User struct {
	ID            string `json:"id"`
	DisplayName   string `json:"displayName"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"EmailVerified"`
}

type UserWithToken struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}
