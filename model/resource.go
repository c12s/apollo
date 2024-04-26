package model

import "fmt"

var (
	RootResource = Resource{resourceId{"", "root"}, nil}
)

type resourceId struct {
	id   string
	kind string
}

type Resource struct {
	id         resourceId
	Attributes []Attribute
}

func NewResource(id, kind string) (*Resource, error) {
	return &Resource{
		id: resourceId{
			id:   id,
			kind: kind,
		},
	}, nil
}

func (r Resource) Id() string {
	return r.id.id
}

func (r Resource) SetId(id string) {
	r.id.id = id
}

func (r Resource) Kind() string {
	return r.id.kind
}

func (r Resource) SetKind(kind string) {
	r.id.kind = kind
}

func (r Resource) Name() string {
	return fmt.Sprintf("%s/%s", r.id.kind, r.id.id)
}
