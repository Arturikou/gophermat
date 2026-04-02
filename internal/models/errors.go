package models

import "errors"

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrLoginNotFound      = errors.New("login not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInsufficientFunds  = errors.New("insufficient funds in the account")
)
