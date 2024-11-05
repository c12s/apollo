package db

import (
	"fmt"
	"os"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
)

type ScyllaManager struct {
	Session *gocqlx.Session
}

func NewScyllaManager() *ScyllaManager {
	return &ScyllaManager{
		Session: CreateSession(),
	}
}

func createCluster() *gocql.ClusterConfig {
	retryPolicy := &gocql.ExponentialBackoffRetryPolicy{
		Min:        time.Second,
		Max:        10 * time.Second,
		NumRetries: 5,
	}

	cluster := gocql.NewCluster(os.Getenv("APOLLO_DB_CLUSTER"))
	cluster.Consistency = gocql.ParseConsistency(os.Getenv("APOLLO_DB_CONSISTENCY"))
	cluster.Keyspace = os.Getenv("APOLLO_DB_KEYSPACE")
	cluster.Timeout = 5 * time.Second
	cluster.RetryPolicy = retryPolicy
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	return cluster
}

func CreateSession() *gocqlx.Session {
	time.Sleep(15 * time.Second) // ovo ceka da se izvrsi init skripta
	cluster := createCluster()
	session, err := gocqlx.WrapSession(cluster.CreateSession())

	if err != nil {
		fmt.Println("An error occurred while creating DB session", err.Error())
		return nil
	}
	return &session
}
