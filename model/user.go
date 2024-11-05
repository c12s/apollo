package model

import (
	"encoding/json"
	"io"
)

type UserDTO struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Org         string `json:"org"`
	Username    string `json:"username"`
	Permissions []string
}

type User struct {
	Name     string `db:"name"`
	Surname  string `db:"surname"`
	Email    string `db:"email"`
	Username string `db:"username"`
}

type Users []*User

func (p *Users) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *User) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(p)
}
