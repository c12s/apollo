package db

import (
	"context"

	"github.com/scylladb/gocqlx/v3"
	"github.com/scylladb/gocqlx/v3/qb"
	"github.com/scylladb/gocqlx/v3/table"
)

// interface za manipulaciju nad tabelama
type ITable interface {
	Metadata() Metadata
	PrimaryKeyCmp() []qb.Cmp
	Name() string

	Get(columns ...string) (stmt string, names []string)
	GetQuery(session *gocqlx.Session, columns ...string) gocqlx.Queryx
	GetQueryContext(ctx context.Context, session *gocqlx.Session, columns ...string) gocqlx.Queryx

	Select(columns ...string) (stmt string, names []string)
	SelectQuery(session *gocqlx.Session, columns ...string) gocqlx.Queryx
	SelectQueryContext(ctx context.Context, session *gocqlx.Session, columns ...string) gocqlx.Queryx
	SelectAll() (stmt string, names []string)

	Insert() (stmt string, names []string)

	Update(columns ...string) (stmt string, names []string)
	UpdateQuery(session *gocqlx.Session, columns ...string) gocqlx.Queryx
	UpdateQueryContext(ctx context.Context, session *gocqlx.Session, columns ...string) gocqlx.Queryx

	Delete(columns ...string) (stmt string, names []string)
}

type Metadata struct {
	M *table.Metadata
}

// za pristup tabelama i rad sa podacima
type Table struct {
	T *table.Table
}

func (t *Table) Metadata() Metadata {
	gocqlxmetadata := t.T.Metadata()

	return Metadata{
		M: &gocqlxmetadata,
	}
}
