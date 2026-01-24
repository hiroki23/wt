# Development

## Requirements

- Go 1.25+
- Git

## Tests

Run all tests:

```
go test ./...
```

## Local Workflow

- `cmd/wt` contains the CLI entrypoint.
- `internal/cli` contains command logic and git helpers.
- Hooks, config parsing, and worktree helpers also live under
  `internal/cli`.

## Release (GoReleaser)

This repo uses GoReleaser to build release artifacts and publish a
Homebrew formula to a tap repo.

### Setup

1) Update `.goreleaser.yaml`:
   - Set `homepage` to the real repository URL.
   - Confirm the tap repo settings.

2) Add GitHub Secrets:
   - `HOMEBREW_TAP_GITHUB_TOKEN`  
     A PAT with write access to the tap repo.

### Release Steps

1) Tag a release:

```
git tag v0.1.0
git push origin v0.1.0
```

2) GitHub Actions runs `.github/workflows/release.yml`:
   - Builds tarballs for supported platforms
   - Publishes GitHub Releases
   - Updates `hiroki23/homebrew-tap` with `Formula/wt.rb`

## Homebrew Tap

`brew install hiroki23/tap/wt`

Homebrew resolves `hiroki23/tap` to the repo
`https://github.com/hiroki23/homebrew-tap`.
