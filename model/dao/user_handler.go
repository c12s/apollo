package dao

import (
	"apollo/model"
	"apollo/model/db"
	"context"
	"log"
	"strings"

	"github.com/scylladb/gocqlx/v3/table"
)

// radi direktno sa user tabelom iz baze
type UserHandler struct {
	manager       *db.ScyllaManager
	userTable     *table.Table
	userEmailView *table.Table
}

func NewUserHandler(manager *db.ScyllaManager) UserHandler {
	return UserHandler{
		manager:       manager,
		userTable:     getUserTable(),
		userEmailView: getUserEmailViewTable(),
	}
}

func (u UserHandler) InsertUser(ctx context.Context, user *model.User) error {
	query := u.manager.Session.Query(u.userTable.Insert()).BindStruct(user)
	if err := query.Exec(); err != nil {
		log.Printf("InsertUserInDb Error: %v", err)
		return err
	}
	log.Printf("user uspesno sacuvan")
	return nil
}

func (h UserHandler) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := h.manager.Session.Query(h.userEmailView.Get()).BindMap(map[string]interface{}{"email": email})
	err := query.GetRelease(&user)

	if err != nil && strings.EqualFold(err.Error(), "not found") {
		log.Printf("user not found in db " + email)
		return nil, err
	} else if err != nil {
		log.Printf("error fetching user by email %v", err.Error())
		return nil, err
	}

	return &user, nil
}

func (h UserHandler) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := h.manager.Session.Query(h.userTable.Get()).BindMap(map[string]interface{}{"username": username})
	err := query.GetRelease(&user)

	if err != nil && strings.EqualFold(err.Error(), "not found") {
		log.Printf("user not found in db " + username)
		return nil, err
	} else if err != nil {
		log.Printf("error fetching user details from db %v", err.Error())
		return nil, err
	}

	return &user, nil
}
