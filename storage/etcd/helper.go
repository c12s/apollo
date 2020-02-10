package etcd

import (
	"strconv"
	"strings"
)

const (
	rbac = "rbac"
)

// keyspace user rbac/namespace:username
// example rbac/default:johndoe
// example data { "name":"john", "lastname": "doe", role:"admin", ...}

// keyspace user resource rbac/namespace:username/resource
// example keyspace rbac/default:johndoe/configs
// example data { "rules" : ["list", "mutate"] }

func userKeyspace(user, namespace string) string {
	temp := strings.Join([]string{user, namespace}, ":")
	return strings.Join([]string{rbac, temp}, "/")
}

func resourceKeyspace(userKeyspace, res string) string {
	return strings.Join([]string{userKeyspace, res}, "/")
}

func delete_empty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func split(data string) []string {
	return delete_empty(strings.Split(data, ","))
}

func ssplit(data, sep string) []string {
	return delete_empty(strings.Split(data, sep))
}

func join(sep string, parts []string) string {
	return strings.Join(parts, sep)
}

func toString(n int64) string {
	return strconv.FormatInt(n, 10)
}
