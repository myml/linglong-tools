# Fuzzy Search Command Design

## Summary

Add a new `linglong-tools fuzzy` subcommand that accepts a keyword as a positional argument
and performs fuzzy search against the remote linglong repository via the existing
`POST /api/v0/apps/fuzzysearchapp` API endpoint.

## Usage

```bash
# Basic keyword search
linglong-tools fuzzy wechat

# With optional filter flags
linglong-tools fuzzy org.deepin.home -c main -a x86_64 -p
```

## Flags

| Flag | Shorthand | Default | Description |
|------|-----------|---------|-------------|
| `--repo` | `-r` | `https://repo.linglong.dev` | Remote repo URL |
| `--name` | `-n` | `stable` | Remote repo name |
| `--channel` | `-c` | `main` | Remote repo channel |
| `--app_arch` | `-a` | `x86_64` | App architecture |
| `--app_version` | `-v` | `""` | App version filter |
| `--prettier` | `-p` | `false` | Pretty-printed JSON output |

## Arguments

- `keyword` (positional, required): The keyword to fuzzy-search app IDs against.

## Implementation

- New file `cmd/fuzzy.go`, following the same structure as `cmd/search.go`
- Keyword is passed to `RequestFuzzySearchReq.AppId` for the API call
- Reuses `initAPIClient()` from `cmd/utils.go` and defaults constants
- Outputs JSON array of matching apps, same format as `search` command
- Build tag: `//go:build !disable_api`

## API Details

- Endpoint: `POST /api/v0/apps/fuzzysearchapp`
- Request body: `RequestFuzzySearchReq` (appId, arch, channel, repoName, version)
- Response: `FuzzySearchApp200Response` with `Data` containing `[]RequestRegisterStruct`