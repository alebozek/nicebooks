package models

import "fmt"

// Representa a un usuario y almacena sus datos relevantes
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// Devuelve una representación en cadena del objeto
func (user User) String() string {
	return fmt.Sprintf("Username: %s, Password: %s, Email: %s", user.Username, user.Password, user.Email)
}
