package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountById(int) (*Account, error)
	GetAccounts() ([]*Account, error)
	Init() error
	createAccountTable() error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	connStr := fmt.Sprintf("user=postgres dbname=postgres password=%s sslmode=disable", dbPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) CreateAccount(a *Account) error {
	query := `INSERT INTO account
    (first_name, last_name, number, balance)
    VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, a.FirstName, a.LastName, a.Number, a.Balance)

	if err != nil {
		return err
	}
	fmt.Println("Account created successfully")
	fmt.Printf("%+v\n", a)
	return nil
}
func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	return nil
}
func (s *PostgresStore) GetAccountById(id int) (*Account, error) {
	return nil, nil
}
func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT * FROM account")
	if err != nil {
		return nil, err
	}
	accounts := []*Account{}
	for rows.Next() {
		a := new(Account)
		err := rows.Scan(&a.ID, &a.FirstName, &a.LastName, &a.Number, &a.Balance, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS account (
		id SERIAL PRIMARY KEY,
		first_name varchar(50),
		last_name varchar(50),
		number SERIAL,
		balance INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}
