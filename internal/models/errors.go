package models

import "errors"

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrLoginNotFound      = errors.New("login not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInsufficientFunds  = errors.New("insufficient funds in the account")
	ErrDuplicateOrder     = errors.New("duplicate order")
	ErrOrderAlreadyExists = errors.New("order already uploaded by this user")
	ErrOrderConflict      = errors.New("order already uploaded by another user")
)
