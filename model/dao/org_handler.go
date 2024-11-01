package dao

import (
	"apollo/model"
	"apollo/model/db"
	"context"
	"log"
	"strings"

	"github.com/scylladb/gocqlx/v3/table"
)

type OrgHandler struct {
	manager  *db.ScyllaManager
	orgTable *table.Table
}

func NewOrgHandler(manager *db.ScyllaManager) OrgHandler {
	return OrgHandler{
		orgTable: getOrgTable(),
		manager:  manager,
	}
}

func (h OrgHandler) FindOrgByName(ctx context.Context, orgName string) (*model.Org, error) {
	var org model.Org
	query := h.manager.Session.Query(h.orgTable.Get()).BindMap(map[string]interface{}{"name": orgName})
	err := query.GetRelease(&org)

	if err != nil && strings.EqualFold(err.Error(), "not found") {
		log.Printf("org not found in db " + orgName)
		return nil, err
	} else if err != nil {
		log.Printf("error fetching org details from db %v", err.Error())
		return nil, err
	}

	return &org, nil
}

func (h OrgHandler) InsertOrg(ctx context.Context, org *model.Org) error {
	query := h.manager.Session.Query(h.orgTable.Insert()).BindStruct(org)
	if err := query.Exec(); err != nil {
		log.Printf("InsertOrg Error: %v", err)
		return err
	}

	return nil
}
