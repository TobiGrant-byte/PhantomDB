package main

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/TobiGrant-byte/phantomdb"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleWrite Role = "write"
	RoleRead  Role = "read"
)

type User struct {
	Username     string
	PasswordHash []byte
	Role         Role
}

func createUser(db *phantomdb.DB, username, password string, role Role) error {
	existing, _ := db.Get([]byte("user:" + username))
	if existing != nil {
		return errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u := User{Username: username, PasswordHash: hash, Role: role}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(u); err != nil {
		return err
	}
	return db.Put([]byte("user:"+username), buf.Bytes())
}

func getUser(db *phantomdb.DB, username string) (User, error) {
	data, err := db.Get([]byte("user:" + username))
	if err != nil {
		return User{}, errors.New("user not found")
	}
	var u User
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&u); err != nil {
		return User{}, err
	}
	return u, nil
}

func verifyPassword(u User, password string) bool {
	err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password))
	return err == nil
}

func listUsers(db *phantomdb.DB) ([]User, error) {
	pairs, err := db.Scan([]byte("user:"), []byte("user:\xff"))
	if err != nil {
		return nil, err
	}
	var users []User
	for _, p := range pairs {
		var u User
		if err := gob.NewDecoder(bytes.NewReader(p.Value)).Decode(&u); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}