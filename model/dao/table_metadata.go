package dao

import (
	"apollo/model/db"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3/table"
)

var (
	userColumns       = []string{"name", "surname", "email", "username"}
	userPartitionKeys = []string{"username"}
	updateuserColumns = []string{"name", "surname"}
	userSortKeys      []string

	orgColumns       = []string{"name", "owner", "members", "permissions"}
	orgPartitionKeys = []string{"name"}
	updateOrgColumns = []string{"owner"}
	orgSortKeys      []string

	userEmailViewColumns   = []string{"name", "surname", "email", "username"}
	userEmailPartitionKeys = []string{"email"}
	updateuserEmailColumns []string
	userEmailSortKeys      []string

	permColumns       = []string{"id", "name"}
	permPartitionKeys = []string{"id"}
	updatePermColumns = []string{"name"}
	permSortKeys      []string
)

const (
	userTableName = "user"
	id            = "id"
	orgTableName  = "org"
	userEmailView = "user_by_email"
	permTableName = "permission"
)

func getTable(tableName string, cols, partitionKeys, sortKeys []string) db.Table {
	tableMeta := table.Metadata{
		Name:    tableName,
		Columns: cols,
		PartKey: partitionKeys,
		SortKey: sortKeys,
	}

	return db.Table{T: table.New(tableMeta)}
}

func getUserTable() *table.Table {
	dbTable := getTable(userTableName, userColumns, userPartitionKeys, userSortKeys)
	return dbTable.T
}

func getOrgTable() *table.Table {
	dbTable := getTable(orgTableName, orgColumns, orgPartitionKeys, orgSortKeys)
	return dbTable.T
}

func getUserEmailViewTable() *table.Table {
	dbTable := getTable(userEmailView, userEmailViewColumns, userEmailPartitionKeys, userSortKeys)
	return dbTable.T
}

func getPermTable() *table.Table {
	dbTable := getTable(permTableName, permColumns, permPartitionKeys, permSortKeys)
	return dbTable.T
}

func genId() string {
	uuid, _ := gocql.RandomUUID()
	return uuid.String()
}
