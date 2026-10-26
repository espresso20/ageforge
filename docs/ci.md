# CI checks

What the repository checks on every pull request and on a schedule, how to run each check on your own machine, and how to quiet a false positive. The tests, `gofmt`, `go vet`, the build and the determinism run live in `go.yml` and `determinism.yml` and are described in [CONTRIBUTING.md](../CONTRIBUTING.md). This page covers the checks added around them.

## Report-only for now

Every check below reports and none can turn a pull request red yet. Each scan step runs with `continue-on-error`, and the findings go to the job summary on the run page or to an artifact. When a check's baseline is clean enough, turn it into a blocking one by removing `continue-on-error: true` from its scan step. To make a check required for merging, add its job name under branch protection in the repository settings.

Baseline on 2026-10-23, so you can tell old findings from new ones:

| Check | Findings |
| --- | ---: |
| govulncheck | 27 called, all in the Go standard library (fixed in go 1.26.1 to 1.26.10) |
| gitleaks | 0 (2 constants allowlisted, see below) |
| CodeQL | 8 in Go (`go/allocation-size-overflow`, `game/account*.go`), 0 in workflows |
| golangci-lint | 2041 (misspell 953, gocritic 688, prealloc 146, perfsprint 74, gosec 60, staticcheck 52, unused 48, unconvert 12, ineffassign 7, errorlint 1) |
| dead code | 211 unreachable functions, 71 when tests count as entry points |
| go.mod hygiene | clean |
| actionlint | 8, all shellcheck remarks in older workflows |
| markdownlint | 174 |
| Vale | 15 |
| typos | 7 |
| lychee | 0 broken links |
| go-licenses | 8 modules, all on the allowlist |
| benchmarks | none (same code on both sides) |

## What runs, and when

| Workflow | Check | Runs on | What it does |
| --- | --- | --- | --- |
| `security.yml` | govulncheck | pull requests, weekly | Known vulnerabilities in the standard library and in the modules the code calls. |
| `security.yml` | gitleaks | pull requests | Secrets in the pull request's commits (the whole history when run by hand). |
| `codeql.yml` | CodeQL | pull requests, master, weekly | GitHub code scanning for Go and for the workflow files. Results are in the Security tab. |
| `scorecard.yml` | OpenSSF Scorecard | master, weekly | Supply-chain score: pinned actions, token permissions, dependency updates. Results are in the Security tab. |
| `lint.yml` | golangci-lint | Go changes, master, weekly | Bug finders, security rules and performance hints (set in `.golangci.yml`). On a pull request it shows only what the change adds. |
| `lint.yml` | go.mod hygiene | Go changes | `go mod tidy` leaves no diff, `go mod verify` passes. |
| `lint.yml` | dead code | weekly, by hand | Functions nothing reaches, as an artifact. |
| `actionlint.yml` | actionlint | changes under `.github/` | Workflow syntax, expressions, and shellcheck on `run:` scripts. |
| `docs.yml` | markdownlint | pull requests, weekly | Markdown structure in the README, CONTRIBUTING, CHANGELOG, `docs/` and `site/docs/`. |
| `docs.yml` | Vale | pull requests, weekly | The wiki's writing rules over `site/docs/` and the README. |
| `docs.yml` | typos | pull requests, weekly | Spelling in code and docs. |
| `docs.yml` | lychee | weekly, by hand | Dead links in `site/docs/` and the README. |
| `licenses.yml` | go-licenses | `go.mod` changes, weekly | License of every built-in module against a permissive allowlist. |
| `bench.yml` | benchstat | Go changes, master | The hot-path benchmarks, base against the change. |
| `nightly.yml` | ordinary path | nightly, by hand | `TestOrdinaryPath` (`smoke/ordinary_path_test.go`) walks the ordinary player through all 22 ages on the frontier with `AGEFORGE_LONG_TESTS=1`. The test skips on pull requests without that variable. |
| `release-snapshot.yml` | GoReleaser | by hand | Builds the release binaries and uploads them as an artifact. Publishes nothing. |

Every workflow sets `permissions: contents: read` at the top and asks for more only on the job that needs it. The workflows that run on pull requests cancel a superseded run of the same pull request. Every action is pinned to a commit SHA with its version in a comment, and Dependabot proposes the bumps (one grouped pull request a week).

## The parked test bot

The smoke suite's test bot is switched off (2026-10-10) and its code builds only with the `smokebot` build tag. CI follows that:

- `go.yml` vets the tagged build in its `check` job (`go vet -tags smokebot ./...`), so the parked code keeps compiling. Build, vet, the test shards, golangci-lint and the dead code report all see the default build, where the bot does not exist.
- Every job that runs `cmd/smoke` passes the tag: the `smoke (fast tier)` job in `go.yml` (label `smoke` or run by hand), the nightly and weekly jobs in `smoke.yml`, and `scripts/determinism.sh`. The Smoke and Determinism workflows are disabled in the repository settings; `gh workflow enable Smoke` and `gh workflow enable Determinism` turn them back on.
- Without the tag `go run ./cmd/smoke` prints how to build it and exits 2, so a job that forgets the tag fails instead of passing on nothing.

To run the bot on your machine, add `-tags smokebot` to the command (`make smoke` does). CONTRIBUTING.md, "Smoke testing", has the details.

## Run a check locally

Pinned versions are in the workflow files. From the repository root:

```bash
# govulncheck
go install golang.org/x/vuln/cmd/govulncheck@v1.9.0 && govulncheck ./...

# golangci-lint (config: .golangci.yml)
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 && golangci-lint run ./...

# go.mod hygiene
go mod tidy -diff && go mod verify

# dead code
go install golang.org/x/tools/cmd/deadcode@v0.52.0 && deadcode -test ./...

# workflow lint
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 && actionlint

# markdownlint
npx markdownlint-cli2@0.23.2 README.md CONTRIBUTING.md CHANGELOG.md "docs/**/*.md" "site/docs/**/*.md"

# dependency licenses
go install github.com/google/go-licenses/v2@v2.0.1
go-licenses report ./... --ignore github.com/espresso20/ageforge

# benchmarks: your working tree against master (needs benchstat)
go install golang.org/x/perf/cmd/benchstat@be2c69fb417e05de5fec05c732911f4c5f172606
scripts/ci-bench.sh compare origin/master /tmp/bench

# release binaries, nothing published (needs goreleaser v2)
goreleaser release --snapshot --clean
```

Vale, typos, lychee and gitleaks ship as release binaries (their module paths or build setups make `go install` a poor fit). Download one from the project's releases page, check it against the release's checksums file, then:

```bash
vale site/docs README.md
typos
lychee --config .lychee.toml site/docs README.md
gitleaks git --redact --log-opts="origin/master..HEAD" .
```

CodeQL and Scorecard run on GitHub only.

## Silence a false positive

Prefer a narrow, commented exception over switching a rule off.

| Check | How |
| --- | --- |
| govulncheck | There is no ignore list. A finding goes away by raising the `go` line in `go.mod` or updating the module. |
| gitleaks | Add an `[[allowlists]]` entry to `.gitleaks.toml` (a path and a line pattern, both must match), or put `gitleaks:allow` in a comment on the line. |
| golangci-lint | `//nolint:gosec // reason` on the line, or a rule under `linters.exclusions` in `.golangci.yml`. |
| actionlint | An `ignore:` pattern under `paths:` in `.github/actionlint.yaml`. |
| markdownlint | `<!-- markdownlint-disable-next-line MD040 -->` above the line, or set the rule to `false` in `.markdownlint.jsonc`. |
| Vale | `<!-- vale AgeForge.Dashes = NO -->` before and `<!-- vale AgeForge.Dashes = YES -->` after, or `<!-- vale off -->` and `<!-- vale on -->`. |
| typos | `spellchecker:disable-line` in a comment on the line, or add the word to `[default.extend-words]` in `_typos.toml`. |
| lychee | Add a pattern to `exclude` in `.lychee.toml`. |
| go-licenses | Add the license to `ALLOWED` in `licenses.yml`, or the module to `--ignore`. |
| CodeQL | Dismiss the alert in the Security tab with a reason. |

## The Vale style

`.vale/styles/AgeForge/` holds five rules and nothing else, so Vale does not grade grammar or run a dictionary:

- `Dashes`: no em or en dashes.
- `Exclamation`: no exclamation marks.
- `USSpelling`: US spelling, from a list of common British forms. "Plough" is left out because The Plough is a tech's name.
- `AgeDuration`: never say how long an age takes ("takes three hours", "about a day", "hours per age").
- `NeverSlow`: never call the game or its pace slow.

## The tools, and why we trust them

Checked on 2026-10-23: the upstream repository is public and open source, has no open security advisory that applies to the pinned version, and has had commits in the last six months unless noted.

| Tool | Pinned | License | Last release | Notes |
| --- | --- | --- | --- | --- |
| govulncheck (golang.org/x/vuln) | v1.9.0 | BSD-3-Clause | 2026-10-11 | The Go team's own tool, `go install`. |
| gitleaks | v8.30.1 (release binary, SHA-256 checked in the workflow) | MIT | 2026-03-21 | The open-source CLI, not the licensed organization action. Last commit 2026-07-22. `go install` fails because the module still declares its old path. |
| CodeQL (github/codeql-action) | v4.39.0 | MIT (action) | 2026-10-15 | GitHub's. The query packs are open source; the CodeQL engine is free on public repositories. An old advisory (fixed in 3.28.3) does not affect v4. |
| OpenSSF Scorecard (ossf/scorecard-action) | v2.5.0 | Apache-2.0 | 2026-10-12 | The OpenSSF's own action. Results are not published to their API. |
| golangci-lint | v2.15.0 (action v9.3.0) | GPL-3.0 (action MIT) | 2026-10-16 | Run as a separate program, nothing is linked into the game. |
| deadcode (golang.org/x/tools) | v0.52.0 | BSD-3-Clause | 2026-10-10 | The Go team's. |
| actionlint | v1.7.12 | MIT | 2026-03-30 | The standard workflow linter. Quiet for six months: last commit 2026-04-19, a mature tool. |
| benchstat (golang.org/x/perf) | commit be2c69fb | BSD-3-Clause | 2026-10-09 | The Go team's. No tagged releases, so a commit is pinned. |
| markdownlint-cli2 (DavidAnson) | action v24.2.0, bundles CLI 0.23.2 | MIT | 2026-08-02 | Maintained by the author of markdownlint. |
| Vale | v3.24.0 (release binary, SHA-256 checked in the workflow) | MIT | 2026-10-01 | The project moved from errata-ai to vale-cli. |
| typos (crate-ci) | v1.51.1 (release binary, SHA-256 checked in the workflow) | Apache-2.0 or MIT | 2026-10-06 | Widely used. |
| lychee | v0.24.2 (lychee-action v2.9.0) | Apache-2.0 or MIT | 2026-05-01 | The action had one low advisory, fixed in 2.0.2. |
| go-licenses (google) | v2.0.1 | Apache-2.0 | 2025-09-08 | Google's. No commit since that release: the weakest pick here. It runs fail-soft and was checked against Go 1.26. |
| GoReleaser | v2.18.3 (action v7.3.0) | MIT | 2026-10-09 | An advisory on secrets in build logs (GHSA-95mm-pj4f-x75r) covers up to 2.18.1 as published. The snapshot job holds no secrets. |
| GitHub's own actions | checkout v7.0.1, setup-go v7.0.0, upload-artifact v7.0.2, download-artifact v8.0.2 | MIT | 2026-07 to 2026-10 | The older workflows keep the majors they used (v4, v5), pinned to the commit those tags pointed at. |

## Release builds

`release.yml` still builds and publishes on a `v*` tag; nothing here changes that. `.goreleaser.yaml` builds the same four targets and the same file names (the in-game updater downloads `ageforge-macos-arm64`, `ageforge-windows-amd64.exe` and so on, plus `SHA256SUMS.txt`), and its own release step is switched off. Run the "Release snapshot" workflow from the Actions tab to get the files as an artifact. To move the tag release over to GoReleaser later, enable `release:` in the config and retire `release.yml` in the same change, so a tag is never released twice.

## Not set up here

These need an owner action in repository settings or an external account:

- The Claude Code Action secret (`CLAUDE_CODE_OAUTH_TOKEN`) for `claude-agent.yml`.
- CodeRabbit (an app install).
- Secret scanning with push protection (repository security settings).
- Branch protection with required checks, to make any of the above blocking.
