# badger

![tag](https://badger-six.vercel.app/circleci/gh/mattsre/badger/pipeline?branch=main&label=tag&value=0.1.$PIPELINE_NUMBER)

Badger is a small self-hosted service for generating GitHub README badges in the [shields.io](https://shields.io/badges) style.

The first supported badge shows the **latest CircleCI pipeline number** for a given branch.

## Quick start

This repo uses [Mise](https://mise.jdx.dev/) for the Go toolchain and tasks, and [Fnox](https://fnox.jdx.dev/) to inject secrets from 1Password at runtime. No secrets are stored in the repository.

```bash
mise install          # install Go 1.26.4 (see mise.toml)
mise run start        # run the server with secrets from fnox.toml
```

Badger listens on `:8080` by default. Override with `BADGER_ADDR` (e.g. `:9090`).

Without Fnox, export the token manually:

```bash
export CIRCLECI_TOKEN=your-personal-api-token   # required for private projects
go run .
```

### Secrets (Fnox)

Secret references live in `fnox.toml` and resolve from 1Password at runtime. Update the provider and vault settings for your environment, and store the actual token in your password manager—not in git.

| Secret            | Description                              |
|-------------------|------------------------------------------|
| `CIRCLECI_TOKEN`  | CircleCI personal API token              |

Create a token at [CircleCI Personal API Tokens](https://app.circleci.com/settings/user/tokens).

## CircleCI pipeline badge

Embed in your README:

```markdown
![pipeline](https://badger.example.com/circleci/gh/myorg/myrepo/pipeline?branch=main)
```

### URL format

```
/circleci/{vcs}/{org}/{repo}/pipeline?branch={branch}
```

| Segment  | Example   | Description                          |
|----------|-----------|--------------------------------------|
| `vcs`    | `gh`      | VCS slug (`gh` for GitHub, `bb` for Bitbucket) |
| `org`    | `myorg`   | Organization or user name            |
| `repo`   | `myrepo`  | Repository name                      |

Query parameters:

| Parameter | Required | Description                          |
|-----------|----------|--------------------------------------|
| `branch`  | yes      | Branch to query (supports `/`, e.g. `feature/foo`) |
| `label`   | no       | Left-side badge label (default: `pipeline`) |
| `value`   | no       | Right-side badge value template; use `$PIPELINE_NUMBER` or `{number}` for the pipeline number (default: pipeline number only) |

Example with a custom label and formatted value:

```markdown
![tag](https://badger.example.com/circleci/gh/myorg/myrepo/pipeline?branch=main&label=tag&value=0.1.$PIPELINE_NUMBER)
```

This renders a badge with `tag` on the left and `0.1.42` on the right when pipeline #42 is the latest on `main`.

`message` is accepted as an alias for `value` (shields.io-style naming).

Example with a custom label only:

```markdown
![build](https://badger.example.com/circleci/gh/myorg/myrepo/pipeline?branch=main&label=build)
```

### Badge colors

The right-side color reflects the pipeline state from CircleCI:

| State                         | Color  |
|-------------------------------|--------|
| success, created              | green  |
| running, pending, setup       | yellow |
| failed, error                 | red    |
| canceled                      | grey   |
| other                         | blue   |

If no pipeline exists for the branch, the badge shows `none` in grey. API failures show `error` in red.

## Configuration

| Variable          | Default  | Description                              |
|-------------------|----------|------------------------------------------|
| `BADGER_ADDR`     | `:8080`  | HTTP listen address                      |
| `CIRCLECI_TOKEN`  | (empty)  | CircleCI personal API token              |

## Health check

```
GET /healthz
```

Returns `200 ok`.

## Development

Requires [Go 1.26](https://go.dev/doc/go1.26) or later. Mise pins the version in `mise.toml`.

```bash
mise install
mise run test         # run tests
go build -o badger .  # build binary
```

Tests do not require a CircleCI token; they use mocks and unit tests only.
