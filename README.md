# jotadrilo/ci

Shared CI for jotadrilo's projects:

- a [Renovate](https://docs.renovatebot.com/) preset, `default.json`, plus an
  opt-in automerge preset, `automerge-dev.json`;
- four reusable GitHub Actions workflows in `.github/workflows/`:
  `go-ci.yml`, `node-ci.yml`, `python-ci.yml` and `monthly-audit.yml`.

Every third-party action is pinned to a full commit SHA with its version in a
comment, every workflow asks for the least permissions it needs, and
`self-test.yml` runs each reusable workflow against the fixtures in
`testdata/` on every push and pull request.

## Renovate preset

In the project's `renovate.json`:

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["github>jotadrilo/ci"]
}
```

What it does:

- extends `config:recommended`, with the Dependency Dashboard issue on;
- runs weekly, before 8am on Monday, Europe/Madrid time;
- groups every minor, patch and digest update of the repository into one PR;
  each major gets its own PR, labeled `major`;
- opens vulnerability fixes at any time, labeled `security` (GitHub's
  Dependabot alerts plus OSV data);
- pins GitHub Actions, reusable workflows included, to commit digests
  (`helpers:pinGitHubActionDigests`) and keeps the SHAs fresh;
- runs `go mod tidy` after Go module updates;
- refreshes npm lockfiles once a month, on the first day of the month;
- range strategy: `auto` overall, `bump` for npm `dependencies` and
  `devDependencies` (apps keep package.json in step with the lockfile),
  `widen` for `peerDependencies`;
- never automerges.

### Opt-in automerge

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["github>jotadrilo/ci", "github>jotadrilo/ci:automerge-dev"]
}
```

`automerge-dev` automerges minor and patch updates of npm `devDependencies`
and minor, patch and digest updates of GitHub Actions, each set in its own
group PR, once the PR's status checks pass. Renovate merges them itself
(`platformAutomerge: false`), so neither GitHub's "Allow auto-merge" setting
nor branch protection is required. It does not automerge a repository with no
status checks at all, so the project needs CI. Order matters: list it after
`github>jotadrilo/ci`.

## Reusable workflows

Pin `@v1` (moves to the latest v1.x.y) or a full commit SHA. With the preset,
Renovate pins `@v1` to its SHA (`@<sha> # v1`) and bumps it when `v1` moves.

### go-ci.yml

Build, vet, test (with `-race` by default), an optional extra test command, an
optional Postgres service, golangci-lint and govulncheck. Two jobs: "Build and
test" and "Lint and vulnerabilities".

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
permissions:
  contents: read
jobs:
  go:
    uses: jotadrilo/ci/.github/workflows/go-ci.yml@v1
```

| Input | Default | |
| --- | --- | --- |
| `working-directory` | `.` | Directory with `go.mod`. |
| `go-version-file` | `go.mod` | Relative to `working-directory`. |
| `go-version` | `""` | Explicit version; overrides the file. |
| `go-latest-patch` | `true` | With `go.mod`, installs the latest patch of its minor (`go 1.26.0` runs on `1.26.x`, with `check-latest`, so a runner image a patch behind does not fall short of the `go` line) so govulncheck sees current stdlib fixes. |
| `run-race` | `true` | `go test -race`. |
| `test-flags` | `""` | Extra `go test` flags, e.g. `-count=1`. |
| `extra-test-command` | `""` | Shell command after `go test`, e.g. `make test-postgres`. |
| `services-postgres` | `false` | Starts Postgres on `localhost:5432` and exports `DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`. |
| `postgres-image` | `postgres:16-alpine` | |
| `golangci-lint-version` | `v2.15.0` | |
| `govulncheck-version` | `v1.9.0` | |
| `run-lint` / `run-govulncheck` | `true` | |

golangci-lint has to be built with a Go at least as new as the `go` line in
`go.mod`, or it fails to load the packages. When a project moves to a new Go
minor, check that `golangci-lint-version` supports it, and raise it with the
caller's `with:` if this repository has not yet.

With Postgres (tests read `DATABASE_URL` and skip without it):

```yaml
jobs:
  go:
    uses: jotadrilo/ci/.github/workflows/go-ci.yml@v1
    with:
      services-postgres: true
      test-flags: -count=1
```

### node-ci.yml

Optional apt packages, `npm ci`, `npm audit --omit=dev`, optional Playwright
Chromium (browsers cached by Playwright version), then lint, test and build.

```yaml
jobs:
  web:
    uses: jotadrilo/ci/.github/workflows/node-ci.yml@v1
    with:
      working-directory: web
      install-playwright: true
```

| Input | Default | |
| --- | --- | --- |
| `working-directory` | `.` | Directory with `package-lock.json`. |
| `node-version-file` | `""` | e.g. `.nvmrc`, relative to `working-directory`; wins over `node-version`. |
| `node-version` | `lts/*` | |
| `install-playwright` | `false` | `playwright` must be a dependency of the project. |
| `apt-packages` | `""` | Space-separated Ubuntu packages installed before `npm ci`, e.g. `librsvg2-bin` for tests that call `rsvg-convert`. |
| `lint-command` | `npm run lint` | Empty skips. |
| `test-command` | `npm test` | Empty skips. |
| `build-command` | `npm run build` | Empty skips. |
| `audit-level` | `high` | `low`, `moderate`, `high` or `critical`; empty skips. |

### python-ci.yml

With [uv](https://docs.astral.sh/uv/), one job per Python version:
`uv sync --locked` (the versions in `uv.lock`), the tests with the project's
`.venv` on `PATH`, [pip-audit](https://pypi.org/project/pip-audit/) on the
runtime dependencies that job installed (hashes required), and `uv build` of
the sdist and the wheel. uv's cache is keyed on `uv.lock` and `pyproject.toml`.
The project needs a `pyproject.toml`, and a `uv.lock` unless `use-lock` is
off.

```yaml
jobs:
  python:
    uses: jotadrilo/ci/.github/workflows/python-ci.yml@v1
    with:
      python-versions: '["3.10", "3.13"]'
```

| Input | Default | |
| --- | --- | --- |
| `python-versions` | `["3.10", "3.13"]` | JSON list; one job each, `fail-fast` off. |
| `working-directory` | `.` | Directory with `pyproject.toml` and `uv.lock`. |
| `use-lock` | `true` | `uv sync --locked`, which also fails when `uv.lock` is out of date with `pyproject.toml`. `false` runs `uv sync --upgrade`: the newest versions the ranges allow, to check the ranges still hold. |
| `sync-args` | `--all-extras` | Extra `uv sync` arguments, e.g. `--extra test` or `--all-groups`. Dependency groups named `dev` are installed by uv anyway. |
| `test-command` | `pytest` | Empty skips. |
| `run-audit` | `true` | pip-audit of `uv export --no-dev --no-emit-project`, the runtime dependencies; extras and dev groups are left out, like `npm audit --omit=dev`. |
| `run-build` | `true` | `uv build`. |
| `uv-version` | `0.13.5` | |
| `pip-audit-version` | `2.10.1` | |

`uv.lock` can pick different versions for different Pythons, so each job
audits what it installed.

### monthly-audit.yml

Runs govulncheck, `npm audit --omit=dev` and pip-audit, lists outdated direct
dependencies (`go list -u -m -json all`, `npm outdated --json`,
`uv tree --outdated --depth 1`), and keeps one
issue per month titled `Monthly dependency audit — YYYY-MM`, vulnerabilities
first. A second run in the same month updates that issue. When a month comes
back clean, no issue is opened, and the open audit issues of earlier months
are closed with a comment. The report is also written to the run's summary.

```yaml
name: Monthly dependency audit
on:
  schedule:
    - cron: "0 6 1 * *"
  workflow_dispatch:
permissions:
  contents: read
  issues: write
jobs:
  audit:
    uses: jotadrilo/ci/.github/workflows/monthly-audit.yml@v1
    with:
      go: true
      node-directories: web
```

| Input | Default | |
| --- | --- | --- |
| `go` | `true` | Audit the Go module. |
| `go-working-directory` | `.` | |
| `govulncheck-version` | `v1.9.0` | |
| `node-directories` | `""` | Space, comma or newline separated; empty skips npm. |
| `node-version` | `lts/*` | |
| `python` | `false` | Audit the Python project: pip-audit of the runtime dependencies locked in `uv.lock`, and its outdated direct dependencies, extras included. |
| `python-working-directory` | `.` | Directory with `pyproject.toml` and `uv.lock`. |
| `python-version` | `3.13` | The Python whose locked versions are audited. |
| `uv-version` | `0.13.5` | |
| `pip-audit-version` | `2.10.1` | |
| `dry-run` | `false` | Summary only; issues untouched. |

A Python project with no Go sets `go: false` and `python: true`.

The caller must grant `issues: write`; the issue is written with the caller's
`GITHUB_TOKEN`.

## Examples

`examples/` holds complete caller files:

- `examples/go-web/`: Go at the root and a web app in `web/` with Playwright
  browser tests (the unpixel-web layout);
- `examples/go-postgres/`: a Go service whose tests need Postgres, plus its
  `web/` app (the flora layout);
- `examples/python-lib/`: a Python package with `uv.lock` at the root (the
  unpixel layout).

Each has `ci.yml`, `monthly-audit.yml` and `renovate.json`.

## Versioning and releases

Tags follow semver: `v1.0.0`, `v1.1.0`, … The major tag `v1` is moved to the
latest `v1.x.y` on every release, so callers on `@v1` get fixes without a
change. A breaking change to an input, a default or a behavior ships as `v2`.

To release, with `main` green:

```sh
git tag -a v1.2.3 -m "v1.2.3"
git tag -fa v1 -m "v1 -> v1.2.3" "v1.2.3^{commit}"
git push origin v1.2.3
git push --force origin v1
```

`^{commit}` makes `v1` point at the commit rather than at the `v1.2.3` tag
object, so the two tags never nest.

This repository's own `renovate.json` extends its presets and has a regex
manager for the tool versions marked with `# renovate:` comments
(golangci-lint, govulncheck, actionlint, renovate, uv, pip-audit), so the defaults stay
current too.

## Notes for private repositories

- GitHub Actions minutes of private repositories count against the account's
  free tier (public repositories run free). The reusable workflows run the Go
  checks as two parallel jobs and the Node checks as one; cancel-in-progress
  on pull requests keeps superseded runs from spending minutes.
- The Renovate GitHub App only opens PRs in repositories it can access. The
  owner grants access to each private repository in GitHub settings
  (Settings → Applications → Renovate → Repository access).
- Vulnerability PRs from GitHub's data need Dependabot alerts enabled on the
  repository (Settings → Code security). OSV data works without it.
