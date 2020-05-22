package vault

import (
	"context"
	"github.com/hashicorp/vault/api"
	"os"
	"strings"
)

type DB struct {
	client *api.Client
}

func New(endpoints []string) (*DB, error) {
	cli, err := api.NewClient(&api.Config{
		Address: endpoints[0],
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		client: cli,
	}, nil
}

func (db *DB) SetToken(token string) {
	db.client.SetToken(token)
}

func (db *DB) Revert() {
	db.client.ClearToken()
}

func (db *DB) Close() {}

func (db *DB) GetToken(ctx context.Context, token string) (map[string]interface{}, error) {
	err := db.unseal()
	if err != nil {
		return nil, err
	}
	defer db.seal()

	db.SetToken(token)
	atoken := db.client.Auth().Token()
	secret, err := atoken.LookupSelf()
	if err != nil {
		return nil, err
	}
	return secret.Data, nil
}

func (db *DB) CreateToken(ctx context.Context, data map[string]string) (string, error) {
	err := db.unseal()
	if err != nil {
		return "", err
	}
	defer db.unseal()

	db.SetToken(os.Getenv("C12S_TOKEN"))
	token := db.client.Auth().Token()
	renew := true
	secret, err := token.Create(&api.TokenCreateRequest{
		Policies:  []string{"default", "c12ssecret"},
		Renewable: &renew,
		TTL:       "1h",
		Metadata:  data,
	})
	if err != nil {
		return "", err
	}
	return secret.Auth.ClientToken, nil
}

func (db *DB) unseal() error {
	tokens := strings.Split(os.Getenv("C12S_KEYS"), ";")
	sys := db.client.Sys()
	for _, token := range tokens {
		_, err := sys.Unseal(token)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) seal() error {
	return db.client.Sys().Seal()
}
