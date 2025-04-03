package auth

import "fmt"

//go:generate easyjson -all -snake_case auth.go

type RequestData struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

func (rd *RequestData) Validate() error {
	if rd.Username == "" {
		return fmt.Errorf("username is required and cannot be empty")
	}
	if rd.Password == "" {
		return fmt.Errorf("password is required and cannot be empty")
	}
	return nil
}
