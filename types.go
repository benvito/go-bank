package main

import "time"

type TransferRequest struct {
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	Note   string `json:"note"`
}

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Number    string `json:"number"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Number    string    `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName, number string) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Number:    number,
		CreatedAt: time.Now().UTC(),
	}
}
