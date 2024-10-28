package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	account, err := NewAccount("John", "Doe", "password1234")
	assert.Nil(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, "John", account.FirstName)
	assert.Equal(t, "Doe", account.LastName)

	fmt.Printf("%+v\n", account)
}
