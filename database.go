package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Database interface {
	CreateAccount(*Account) (*Account, error)
	UpdateAccount(*Account) error
	DeleteAccount(int) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
}

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB() (*PostgresDB, error) {
	connectionString := "user=postgres dbname=gobank password=12345 sslmode=disable"
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{db: db}, nil
}

func (db *PostgresDB) Init() error {
	return db.createAccountTable()
}

func (db *PostgresDB) createAccountTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			first_name VARCHAR(50) NOT NULL,
			last_name VARCHAR(50) NOT NULL,
			number VARCHAR(50) NOT NULL UNIQUE,
			balance int,
			created_at TIMESTAMP
		)
	`

	_, err := db.db.Exec(query)
	return err
}

func (db *PostgresDB) CreateAccount(acc *Account) (*Account, error) {
	query := `
	INSERT INTO accounts
	(first_name, last_name, number, balance, created_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *
	`
	rows, err := db.db.Query(query, acc.FirstName, acc.LastName, acc.Number, acc.Balance, acc.CreatedAt)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanRowToAccount(rows)
	}

	return nil, fmt.Errorf("account was not inserted")
}

func (db *PostgresDB) UpdateAccount(acc *Account) error {
	return nil
}

func (db *PostgresDB) DeleteAccount(id int) error {
	_, err := db.db.Query("DELETE FROM accounts WHERE id = $1", id)
	return err
}

func (db *PostgresDB) GetAccountByID(id int) (*Account, error) {
	rows, err := db.db.Query("SELECT * FROM accounts WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanRowToAccount(rows)
	}

	return nil, fmt.Errorf("accound with ID=%d not found", id)
}

func (db *PostgresDB) GetAccounts() ([]*Account, error) {
	rows, err := db.db.Query("SELECT * FROM accounts")

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanRowToAccount(rows)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanRowToAccount(rows *sql.Rows) (*Account, error) {
	account := &Account{}

	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
	)

	return account, err
}
