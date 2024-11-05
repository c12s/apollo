package dao

import (
	"apollo/model"
	"apollo/model/db"
	"context"
	"log"

	"github.com/scylladb/gocqlx/v3/table"
)

// radi direktno sa org tabelom iz baze
type PermHandler struct {
	manager   *db.ScyllaManager
	permTable *table.Table
}

func NewPermHandler(manager *db.ScyllaManager) PermHandler {
	return PermHandler{
		permTable: getPermTable(),
		manager:   manager,
	}
}

func (h PermHandler) GetAllPerms(ctx context.Context) ([]string, error) {
	var perms []model.Permission
	query := h.manager.Session.Query(h.permTable.SelectAll())
	err := query.SelectRelease(&perms)
	if err != nil {
		log.Printf("error fetching all perms from db %v", err)
		return nil, err
	}
	return getPermNames(perms), nil
}

func getPermNames(perms []model.Permission) []string {
	permNames := make([]string, len(perms))
	for i, perm := range perms {
		permNames[i] = perm.Name
	}
	return permNames
}
