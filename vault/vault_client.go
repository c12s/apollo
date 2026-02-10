package vault

import (
	"apollo/model"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	vault "github.com/hashicorp/vault-client-go"
	schema "github.com/hashicorp/vault-client-go/schema"
	"go.opentelemetry.io/otel"
)

type VaultClientService struct {
	client *vault.Client
}

type VaultKey struct {
	RootKey   string `json:"root_key"`
	UnsealKey string `json:"unseal_key"`
}

// init
func NewVaultClientService(ctx context.Context) (*VaultClientService, error) {
	return &VaultClientService{
		client: initClient(ctx),
	}, nil
}

func initClient(ctx context.Context) *vault.Client {
	vaultAddress := fmt.Sprintf("http://%s:%s",
		os.Getenv("VAULT_HOSTNAME"),
		os.Getenv("VAULT_HTTP_PORT"))

	client, err := vault.New(
		vault.WithAddress(vaultAddress),
		vault.WithRequestTimeout(30*time.Second),
	)

	if err != nil {
		log.Printf("Error creating vault client %v", err)
	}

	// check if its initialized
	initResp, err := client.System.ReadInitializationStatus(context.Background())
	if err != nil {
		log.Printf("Init status error: %v", err)
	}

	initStatus := initResp.Data["initialized"].(bool)

	if initStatus {
		log.Println("Vault already initialized.")
		vaultKey := loadKeyFromJson()
		if err := client.SetToken(vaultKey.RootKey); err != nil {
			log.Printf("Error while trying to set vault token: %v", err)
		}

		Unseal(ctx, client, vaultKey.UnsealKey)
		return client
	}

	// init
	initializedVault := Initialize(ctx, client)
	vaultKey := VaultKey{
		RootKey:   initializedVault.rootKey,
		UnsealKey: initializedVault.keysArray[0].(string),
	}
	saveKeyToJson(vaultKey)

	// auth
	if err := client.SetToken(initializedVault.rootKey); err != nil {
		log.Fatal(err)
	}

	Unseal(ctx, client, vaultKey.UnsealKey)
	MountSecretEngine(ctx, client)

	return client
}

func (v VaultClientService) RegisterUser(ctx context.Context, username string, password string, policies []string) {
	tracer := otel.Tracer("vault-client")
	ctx, span := tracer.Start(ctx, "Vault.RegisterUser")
	defer span.End()

	resp, err := v.client.Auth.UserpassWriteUser(
		ctx,
		username,
		schema.UserpassWriteUserRequest{
			Password:    password,
			Policies:    policies,
			TokenPeriod: "0.5h",
		},
		vault.WithMountPath("userpass"),
	)
	if err != nil {
		log.Println("vault registration failed")
		log.Printf("Error: %v", err)
	} else {
		log.Println("vault registration finished")
		log.Println(resp)
	}

}

func (v VaultClientService) LoginUser(ctx context.Context, req model.LoginReq) model.LoginResp {
	tracer := otel.Tracer("vault-client")
	ctx, span := tracer.Start(ctx, "Vault.LoginUser")
	defer span.End()

	resp, err := v.client.Auth.UserpassLogin(
		ctx,
		req.Username,
		schema.UserpassLoginRequest{
			Password: req.Password,
		},
		vault.WithMountPath("userpass"),
	)
	if err != nil {
		log.Printf("VaultLogin error: %v", err)
		return model.LoginResp{Token: "", Error: err}
	}

	return model.LoginResp{Token: resp.Auth.ClientToken, Error: nil}
}

func (v VaultClientService) VerifyToken(ctx context.Context, token string) model.VerificationResp {
	resp, err := v.client.Auth.TokenLookUp(
		ctx,
		schema.TokenLookUpRequest{
			Token: token,
		},
	)

	if err != nil {
		log.Printf("%v", err)
		return model.VerificationResp{Verified: false, Username: ""}
	}

	expTime := resp.Data["expire_time"].(string)
	metaMap := resp.Data["meta"].(map[string]interface{})
	username := metaMap["username"].(string)
	timestamp, err := time.Parse(time.RFC3339Nano, expTime)

	if err != nil {
		log.Printf("Error parsing timestamp: %v", err)
		return model.VerificationResp{Verified: false, Username: username}
	}

	currentTime := time.Now()
	isBefore := timestamp.Before(currentTime)

	return model.VerificationResp{Verified: !isBefore, Username: username}
}

func Initialize(ctx context.Context, client *vault.Client) VaultClient {
	tracer := otel.Tracer("vault-client")
	ctx, span := tracer.Start(ctx, "Vault.Initialize")
	defer span.End()

	resp, err := client.System.Initialize(
		ctx,
		schema.InitializeRequest{
			PgpKeys:         nil,
			RootTokenPgpKey: "",
			SecretShares:    1,
			SecretThreshold: 1,
		},
	)
	if err != nil {
		log.Printf("Vault failed to initialize %v", err)
	}

	keysArray, ok := resp.Data["keys"].([]interface{})
	if !ok || len(keysArray) == 0 {
		log.Println("Error: Unable to access the 'keys' array")
		return VaultClient{}
	}

	return VaultClient{keysArray: keysArray, rootKey: resp.Data["root_token"].(string)}
}

func Unseal(ctx context.Context, client *vault.Client, firstKey string) {
	tracer := otel.Tracer("vault-client")
	ctx, span := tracer.Start(ctx, "Vault.Unseal")
	defer span.End()

	_, err := client.System.Unseal(
		ctx,
		schema.UnsealRequest{
			Key: firstKey, // first key in array
		},
	)
	if err != nil {
		log.Printf("Vault failed to unseal: %v", err)
	}
}

func MountSecretEngine(ctx context.Context, client *vault.Client) {
	tracer := otel.Tracer("vault-client")
	ctx, span := tracer.Start(ctx, "Vault.MountSecretEngine")
	defer span.End()

	_, err := client.System.AuthEnableMethod(
		ctx,
		"userpass",
		schema.AuthEnableMethodRequest{
			Description: "Mount for user identity",
			Type:        "userpass",
		},
	)
	if err != nil {
		log.Printf("Vault failed to mount secret engine %v", err)
	}
}

var path = os.Getenv("VAULT_KEYS_FILE")

func loadKeyFromJson() VaultKey {
	jsonFile, err := os.ReadFile(path)
	if err != nil {
		log.Printf("%s", err)
	}

	var vaultKey VaultKey
	err = json.Unmarshal(jsonFile, &vaultKey)
	if err != nil {
		log.Println("Error:", err)
	}

	return vaultKey
}

func saveKeyToJson(vaultKey VaultKey) {
	updatedJSON, err := json.MarshalIndent(vaultKey, "", "  ")
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	err = ioutil.WriteFile(path, updatedJSON, 0644)
	if err != nil {
		log.Println("Error writing JSON file:", err)
		return
	}

}
