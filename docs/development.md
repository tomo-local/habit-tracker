# Development Guide

For contributors building and running habit-tracker from source. For install/usage as an end user, see the [README](../README.md).

## Google Cloud OAuth setup (local development)

Release binaries ship with a default OAuth client baked in at build time (see `internal/config/env.go`, injected via `-ldflags -X ...defaultClientID=... -X ...defaultClientSecret=...` in `.github/workflows/release.yml`). A local `go build`/`go run` has empty defaults, so you need your own OAuth client to authenticate during development.

1. Create a project in the [GCP Console](https://console.cloud.google.com/).
2. Under "APIs & Services" → "Library", enable the **Google Calendar API**.
3. Under "APIs & Services" → "OAuth consent screen", set User Type to **External** and add your own Google account as a test user.
4. Under "Credentials" → "Create Credentials" → "OAuth client ID", choose application type **Desktop app**.
5. Copy the generated Client ID and Client Secret.
6. Copy `.envrc.example` to `.envrc` and fill in `HABIT_CLIENT_ID` / `HABIT_CLIENT_SECRET` with those values.

`internal/config.Config.ClientID()` / `ClientSecret()` read `HABIT_CLIENT_ID` / `HABIT_CLIENT_SECRET` first and only fall back to the (empty, in source) embedded defaults — so a local build always uses your own client, never the release one.

```sh
go run . auth login   # opens the OAuth consent screen using your client
```

## Directory structure

```
.
├── main.go                # entry point: parses the subcommand and dispatches to cmd/
├── cmd/                    # one file per subcommand, wired together in cmd.go
│   ├── cmd.go              # Cmd struct + shared dependencies (config, calendar, prompter)
│   ├── auth.go             # `auth login` — runs the OAuth flow, saves token.json
│   ├── setup.go            # `setup` — create/select a calendar, register habit names
│   ├── add.go              # `add` — record today's habit as a calendar event
│   └── view.go             # `view` (and the no-arg default) — render the heatmap
├── internal/
│   ├── auth/               # OAuth authorization: local callback server + code exchange
│   ├── calendar/           # Google Calendar API client (list/create calendar, add/list events)
│   ├── config/             # config.json, token.json, and env var handling (HABIT_*)
│   ├── heatmap/            # renders the GitHub-contributions-style heatmap
│   └── prompt/             # interactive prompt interface, with a huh/ implementation
├── pages/                  # static GitHub Pages site (home page, privacy policy)
├── Formula/                # Homebrew Formula (see brew install in the README)
├── scripts/                # release/CI helper scripts
├── docs/                   # documentation (this file, i18n README)
├── mise.toml               # Go version, `mise run build` / `mise run test` tasks
└── .envrc.example          # template for the local OAuth env vars above
```

## Build & test

```sh
mise run build   # go build -o bin/habit
mise run test    # go test ./...
```

## Release

Releases are cut by the repository owner by pushing a `v*` tag; `.github/workflows/release.yml` handles the rest (cross-compiling binaries, publishing the GitHub Release with notes generated from Conventional Commits messages since the previous tag, and updating `Formula/habit-tracker.rb`). Contributors don't need to do anything for this.
