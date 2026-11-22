# apollo — c12s identity service

The identity provider of the control plane: a gRPC `AuthService` (port
`8002`) that owns user/org/permission records in Scylla, keeps passwords in
**Vault** (userpass), publishes policy relationships to **Oort**, and mints
the JWT carrying the `permissions` claim that every protected service checks.

## Role in c12s

```
cockpit register/login ──▶ gateway ──▶ apollo ──▶ Scylla (users, orgs, permission catalog)
                                            ├──▶ Vault  (passwords, tokens, api_key.json)
                                            └──▶ Oort   (org→user inheritance, policies)

every protected request: gateway ──▶ apollo.VerifyToken ──▶ Vault (is token valid?)
                                                   └──▶ Oort (granted permissions)
           ← JWT {username, permissions} → forwarded as gRPC metadata `authz-token`
```

`VerifyToken` sits on the **hot path of every authenticated request**: the
gateway calls it, gets a freshly minted JWT whose `permissions` are the
(current) permissions Oort grants to the user, and forwards that JWT to
Magnetar/Gravity/etc. This is why an Oort outage takes down the whole API.

## API (`proto1/service.proto`, gRPC reflection enabled)

| RPC | Behavior |
|---|---|
| `RegisterUser` | creates the Scylla user, the org (with the full permission catalog as owner permissions), the Oort org→user inheritance, async Oort policies (NATS), and the Vault userpass user with policy `org.add` |
| `LoginUser` | Vault userpass login → returns a Vault **client token** (period `0.5h`) |
| `VerifyToken` | validates the presented token against Vault, fetches granted permissions from Oort, mints an `HS256` JWT — claims `username`, `permissions` (`perm|kind|id,` list), `exp` +24h |
| `DecodeJwt` | parses that JWT → permission strings |
| `Authorize` | **stub that always returns `authorized: true`** |

Registration request fields (via gateway `POST /apis/core/v1/users`):
`email`, `username` (unique, alphanumeric + `_ - .`), `name`, `surname`,
`password`, `org`. Registration fails when the org, username, or email
already exists.

## Running

Part of the control-plane compose stack (`tools/docker-compose.yml`,
service `apollo`). It needs, in order: Scylla with the schema applied
(`tools/start.sh` does both), Vault reachable, Oort reachable for
`VerifyToken`.

```bash
docker compose up -d apollo        # from tools/
# not runnable standalone outside the compose network — see hardcoded addresses below
```

## Configuration (env)

Only these variables are actually read by the code:

| Variable | Used for |
|---|---|
| `VAULT_HOSTNAME`, `VAULT_HTTP_PORT` | Vault address (`http://$host:$port`) |
| `VAULT_KEYS_FILE` | path of the key file Apollo writes (`/etc/apollo/api_key.json`, bind-mounted from `tools/vault/api_key.json`) |
| `SECRET_KEY` | JWT signing key (`HS256`), shared with all services that check `permissions` |
| `APOLLO_DB_CLUSTER`, `APOLLO_DB_CONSISTENCY`, `APOLLO_DB_KEYSPACE` | Scylla connection (`apollo` keyspace) |

The compose file also passes `IAM_ADDRESS`, `OORT_ADDRESS`,
`MAGNETAR_ADDRESS`, `NATS_*` — **the code ignores them**; the port is
hardcoded to `8002` and the Oort/NATS clients to `oort:8000` / `nats:4222`.

## Data and dependencies

- **Scylla** (`apollo` keyspace): `user`, `org`, `permission` tables plus the
  `user_by_email` materialized view. The permission catalog rows are inserted
  by `tools/scylla/create_schema_scylla.cql`.
- **Vault**: userpass users (token period 0.5h) + root/unseal keys in
  `api_key.json` (Apollo initializes and unseals Vault on first start).
- **Oort**: inheritance relationship `org → user` plus one policy per
  permission, submitted asynchronously over NATS.
- **Scylla availability = login/registration availability**; **Oort
  availability = authorization availability**.
