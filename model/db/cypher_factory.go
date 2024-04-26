// TO DO
package db

import (
	"apollo/model"
)

type CypherFactory interface {
	createResource(req model.User) (string, map[string]interface{})
	findUser(req model.LoginReq) (string, map[string]interface{})
	getUsers() (string, map[string]interface{}) 
}

type simpleCypherFactory struct {
}

func NewSimpleCypherFactory() CypherFactory {
	return &simpleCypherFactory{}
}

const ncCreateResourceCypher = `
CREATE (u:User {id: $id, name: $name, surname: $surname, email: $email, password: $password, org: $org, permission: $permission, username: $username}) RETURN u
`

func (f simpleCypherFactory) createResource(req model.User) (string, map[string]interface{}) {
	return ncCreateResourceCypher,
		map[string]interface{}{
			"id":       req.Id,
			"name":     req.Name,
			"surname":     req.Surname,
			"email":     req.Email,
			"password":     req.Password,
			"org":     req.Org,
			"username": req.Username,
}
}

const ncGetUsers = `
MATCH (u:User)
RETURN u
`
func (f simpleCypherFactory) getUsers() (string, map[string]interface{}) {
	return ncGetUsers,
		nil
}

const ncFindUser = `
MATCH (user:User)
WHERE user.username = $username AND user.password = $password
RETURN user AS user, user.name AS name
`
func (f simpleCypherFactory) findUser(req model.LoginReq) (string, map[string]interface{}) {
	return ncFindUser,
		map[string]interface{}{
			"username":      req.Username,
			"password":   req.Password,
}
}



