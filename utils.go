package main

import (
	"fmt"
	"os"
	"strings"

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

func SplitMessage(message string, chunkSize int) []string {
	if len(message) <= chunkSize {
		return []string{message}
	}

	var parts []string
	for len(message) > chunkSize {
		// Find the last space within the chunk to avoid cutting words
		i := strings.LastIndex(message[:chunkSize], " ")
		if i == -1 {
			i = chunkSize
		}

		parts = append(parts, message[:i])
		message = message[i:]
	}

	if len(message) > 0 {
		parts = append(parts, message)
	}

	return parts
}
