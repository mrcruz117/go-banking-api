package main

import "log"

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	// seedAccounts(store)

	server := NewAPIServer(":8088", store)
	server.Run()
}

func seedAccount(s Storage, firstName, lastName, password string) *Account {
	account, err := NewAccount(firstName, lastName, password)
	if err != nil {
		log.Fatal(err)
	}
	err = s.CreateAccount(account)
	if err != nil {
		log.Fatal(err)
	}
	return account
}

func seedAccounts(s Storage) {

	seedAccount(s, "Michael", "Cruz", "strongPassword123")
	seedAccount(s, "Sophia", "Cruz", "strongPassword1234")
	seedAccount(s, "John", "Doe", "password1")
	seedAccount(s, "Jane", "Doe", "password2")
	seedAccount(s, "Jim", "Beam", "password3")
	seedAccount(s, "Jack", "Daniels", "password4")
	seedAccount(s, "Chris", "Pines", "password5")
	seedAccount(s, "Anthony", "Gonzales", "password6")
}
