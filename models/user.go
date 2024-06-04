package models

import "fmt"

type User struct {
	ID       uint
	Username string
	Password string
	Email    string
}

func (user User) String() string {
	return fmt.Sprintf("Username: %s, Password: %s, Email: %s", user.Username, user.Password, user.Email)
}
