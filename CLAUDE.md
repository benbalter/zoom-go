# CLAUDE.md

`zoom` is a Go CLI that prints your next Google Calendar event and, if it's a Zoom meeting starting soon, opens it in the Zoom app. It's a port of the Ruby [zoom-launcher](https://github.com/benbalter/zoom-launcher). See the [README](README.md) for install and first-run authorization.

## Commands

- `script/cibuild` is what CI runs ([ci.yml](.github/workflows/ci.yml)): `go mod download` then `go test ./...`.
- `go vet ./...` is clean. `gofmt -l .` flags only [`cmd/zoom/main.go`](cmd/zoom/main.go), whose doc comment predates the Go 1.19 comment format. Run `gofmt -w` on files you touch.
- `go build ./...` to compile. [`script/build`](script/build) writes a `zoom-launcher` binary to the repo root, which `.gitignore` doesn't cover; don't commit it.
- Don't run the binary against a real account in a session: it reads OAuth credentials from `~/.config/google` and opens the browser and Zoom.

## Architecture

- [`cmd/zoom/main.go`](cmd/zoom/main.go): flags (`-import`), setup instructions, the OAuth flow, and printing the meeting.
- [`zoom.go`](zoom.go): fetching upcoming events and deciding whether one is "soon". [`call.go`](call.go): pulling a Zoom URL out of an event and turning it into a `zoommtg://` app URL.
- [`client.go`](client.go): Google OAuth client and Calendar service. [`config/`](config/): the `Provider` interface and the file-backed implementation, shared with the Ruby gem's credential location.
- Tests in [`zoom_test.go`](zoom_test.go) use a fake Calendar API over `httptest`, so they need no credentials or network.

## Releasing

[`script/release`](script/release) creates an annotated tag and runs GoReleaser to publish binaries. That's outward-facing: only run it with the owner's explicit go-ahead. The last release was in 2018, so check [`.goreleaser.yml`](.goreleaser.yml) (`archive`, `replacements`) and the script's `--rm-dist` flag against current GoReleaser (`goreleaser check`) before the next one.

## Gotchas

- There's no `master` branch; the default is `main`. Install from source with `go install github.com/benbalter/zoom-go/cmd/zoom@main`, since `@latest` resolves to the 2018 `v0.2.0` tag.
- [`.travis.yml`](.travis.yml) is dead config; CI is GitHub Actions.
- Dependency bumps arrive as Renovate PRs (shared `benbalter/renovate-config:instant` preset) that automerge when CI passes.
