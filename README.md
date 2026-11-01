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
Magnetar/Gravity/etc. This is why an Oort outage takes down the whole API
(see bugs).

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

## Status

Stable for the lab auth chain (every protected request depends on it); the
robustness findings below are accepted lab risks until a v2 hardening pass.

## Bugs, disabled and unfinished features

### Confirmed bugs

| Where | Effect |
|---|---|
| `client/oort_client.go` — `GetGrantedPermissions` | Uses `log.Fatalln(err)` on RPC error: **if Oort is unreachable while `VerifyToken` runs, the Apollo process exits.** A CP component outage thus crashes the identity service instead of returning an error. Same pattern (`log.Fatal` on dial) in `AuthorizeUser`. |
| Hardcoded endpoints | `oort:8000`, `nats:4222`, port `8002` are constants; the compose env vars for them are ignored. Apollo cannot run outside the compose network. |
| `server/auth_server.go` — `Authorize` | Returns `{authorized: true}` unconditionally — an unimplemented stub exposed on the gRPC surface. (The real check is Oort's `Authorize`; nothing in the lab routes to this RPC.) |
| `repository/user_repo_impl.go` — registration | **Not transactional**: Scylla user is inserted before the org row; a later failure (org insert, Oort relationship) leaves a half-registered user whose re-registration then fails with "user already exists". Async policy submission failures are only logged. |
| Orgs get **all** permissions | A new org owner receives the entire permission catalog; combined with the stub above, the lab model is permissive by default. Acceptable for experiments, needs redesign for multi-tenant use. |

### Disabled / unfinished features

- `configs/nats` configuration is never consumed (dead code).
- `repository.UserRepo.LoginUser` is a stub (`invalid mapping`) — login is
  implemented directly against Vault.
- No tests (`go test ./...` → no test files).
- Vault is fully trusted for password storage, but **the root token is stored
  on disk and used for all operations**; userpass tokens (0.5h) and the issued
  JWT (24h) have unrelated lifetimes.

### PoC boundaries (by design, not bugs)

- Plain HTTP to Vault inside the compose network; `SECRET_KEY` is a shared
  lab value and every service trusts it.
- `RegisterUser` only checks username/email/org uniqueness (no email
  verification, no password policy).
- The permission snapshot in the JWT is taken at request time via
  `VerifyToken`, so Oort policy changes apply from the next request — a
  deliberate simplification (no token revocation list).

References: `../plan/architecture.md` §auth chain, `../tools/docs/README.md`
(corrected API reference), `../plan/review.md` (S1: NATS trust boundary).
