# GizClaw E2E

This directory contains the manual/setup-driven GizClaw e2e suites. These tests
depend on a prepared local e2e server, shared resource fixtures, and generated
CLI config homes. Go test files in this tree require the `gizclaw_e2e` build tag
so they are not pulled into ordinary `go test ./...` runs.

## Modules

- `testdata/`: committed e2e fixtures plus ignored generated runtime files.
- `setup/`: lifecycle scripts for building the CLI, starting services,
  resetting shared data, and stopping services.
- `client/`: typed client and protocol-level e2e tests.
- `cmd/`: user-facing `gizclaw` CLI command e2e tests.
- `ui/`: browser UI e2e tests for Admin UI, Play UI, and cross-surface smoke
  checks.

## Standard Flow

1. Copy `test/gizclaw-e2e/.env.example` to `test/gizclaw-e2e/.env`, then fill
   provider credential values. Runtime addresses, resource names, resource IDs,
   model IDs, voice IDs, and e2e identity keys are committed fixtures, not env
   values.

2. Build the e2e CLI binary:

```sh
./test/gizclaw-e2e/setup/build.sh
```

3. Start the local e2e server:

```sh
./test/gizclaw-e2e/setup/start-server.sh
```

4. Clear and initialize shared server data:

```sh
./test/gizclaw-e2e/setup/reset_data.sh
```

5. Run the needed client or CLI tests, for example:

```sh
go test -tags gizclaw_e2e -count=1 ./test/gizclaw-e2e/client/workspace
go test -tags gizclaw_e2e -count=1 ./test/gizclaw-e2e/client/rpc
go test -tags gizclaw_e2e -count=1 ./test/gizclaw-e2e/cmd/connect
```

6. For browser UI tests, start the matching UI surface first:

```sh
./test/gizclaw-e2e/setup/start-admin-ui.sh
./test/gizclaw-e2e/setup/start-play-ui.sh
```

Then run the relevant UI suite:

```sh
go test -tags gizclaw_e2e -count=1 ./test/gizclaw-e2e/ui/admin/...
go test -tags gizclaw_e2e -count=1 ./test/gizclaw-e2e/ui/play/...
go test -tags gizclaw_e2e -count=1 ./test/gizclaw-e2e/ui/smoke/...
```

7. Stop e2e services when finished:

```sh
./test/gizclaw-e2e/setup/stop.sh
```

## Test Data

`testdata/resources` is the shared source for server resources used by client,
cmd, and UI tests. Resource fixture filenames use a three-digit numeric prefix,
for example `000-openai-credential.json` or
`040-workflow-flowcraft-chat.json`.

Only credential-like provider values should be environment placeholders, such as
`${GIZCLAW_E2E_OPENAI_API_KEY}`. Values are supplied by
`test/gizclaw-e2e/.env` during setup. `reset_data.sh` skips real-provider
fixtures whose required credentials are empty, while still initializing shared
fixtures with committed non-secret defaults. Do not commit real provider keys,
tokens, app secrets, or access keys. Stable e2e identity key pairs are committed
config fixtures, not env values.

Generated runtime data under `testdata/server-workspace/data/` and generated
binaries under `testdata/bin/` stay ignored.

## Config Homes

`testdata/admin-config-home` and `testdata/gizclaw-config-home` are
`XDG_CONFIG_HOME` roots. They must contain the normal `gizclaw/` config layout
and committed `identity.key` fixtures. Do not hand-maintain derived public keys
in committed config files; use `identity-key` or private-key fields and let the
config loaders derive public keys at runtime.

## Client Tests

`client/rpc` contains typed RPC coverage. Test files should be grouped by RPC
module prefix, and individual methods should be split by `Test...` functions.

`client/workspace` contains workspace voice and history cases as ordinary
`_test.go` files. It should not use a custom `main.go -case ...` dispatcher.

`client/social` contains friend and friend-group behavior. These tests are
client-driven, not CLI-story driven, and should cover relation changes,
workspace ACL visibility, message rounds, `workspace.history.updated`, history
list/get cursor behavior, and history replay.

## CLI Tests

`cmd` tests run the real `gizclaw` binary from `testdata/bin/gizclaw` through
Go `os/exec`. They should not use `go run` and should not shortcut through
typed clients.

The `cmd` layout mirrors the real CLI command hierarchy: `root`, `gen-key`,
`context`, `serve`, `service`, `migrate`, `connect`, `admin`, and `play`. Each
command directory has one `USER_STORIES.md` plus focused `_test.go` files.

`cmd/play` tests the `gizclaw play` command. Browser Play UI behavior belongs
under `ui/play`.

## UI Tests

`ui/admin` and `ui/play` are browser UI tests organized by visible page or major
surface. Missing shared resources should be added under `testdata/resources` and
initialized by `setup/reset_data.sh`.

`ui/smoke` contains cross-surface checks, such as opening both Admin UI and Play
UI against the same seeded test service.
