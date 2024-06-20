package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func CreateDirecryIfNotExist(directory string) error {
	// check if directory exist
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// try creating a new directory
		if err := os.Mkdir(directory, 0755); err != nil {
			fmt.Println("Error creating directory for database", err)
			return err
		} else {
			fmt.Println("New directory for database created!")
			return nil
		}
	} else {
		return nil
	}
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ContainsAlert(alerts []Alert, id string) bool {
	for _, alert := range alerts {
		if alert.Id == id {
			return true
		}
	}
	return false
}
