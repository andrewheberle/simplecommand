# AGENTS.md

Guidance for AI coding agents and contributors working in this repository.

## Repository layout

This repository contains **three separate Go modules**:

| Path | Module | Notes |
|------|--------|-------|
| `/` | `github.com/andrewheberle/simplecommand` | Core `Command` type. Minimal dependencies. Do not add `viper` or `koanf` here. |
| `vipercommand/` | `github.com/andrewheberle/simplecommand/vipercommand` | Viper-enabled `Command`. Has its own `go.mod` and `go.sum`. |
| `koanfcommand/` | `github.com/andrewheberle/simplecommand/koanfcommand` | koanf-enabled `Command`. Has its own `go.mod` and `go.sum`. |

`vipercommand` and `koanfcommand` depend on a **published** version of
`simplecommand` (see their `go.mod` files), not the local copy. Do not add a
`replace` directive. Users of these modules ignore `replace` directives, so
testing against the local copy would hide breakage that users would then hit.

To work on the modules together, create a `go.work` file in the repository
root. It is in `.gitignore` and must not be committed:

```sh
go work init . ./vipercommand ./koanfcommand
```

While a `go.work` file exists, Go builds `vipercommand` and `koanfcommand`
against the local `simplecommand`. Before committing, also run their tests
with `GOWORK=off` to check them against the published version, which is what users
get.

### Releases

Releases are managed by [release-please](https://github.com/googleapis/release-please)
(`.github/workflows/release-please.yml`). **Do not create tags or GitHub
releases by hand, and do not edit `CHANGELOG.md` files.**

After each push to `main`, release-please reads the Conventional Commit
messages since the last release of each module. It then opens or updates a
separate release PR for each module that has releasable changes, with the next
version and changelog. Merging a release PR creates the tag and GitHub
release:

| Module | Tag | Changelog |
|--------|-----|-----------|
| `simplecommand` | `vX.Y.Z` | `CHANGELOG.md` |
| `vipercommand` | `vipercommand/vX.Y.Z` | `vipercommand/CHANGELOG.md` |
| `koanfcommand` | `koanfcommand/vX.Y.Z` | `koanfcommand/CHANGELOG.md` |

- `fix:` gives a patch release, and `feat:` a minor release.
- Breaking changes (`!`) give a minor release while a module's version is
  below 1.0.0, and a major release after that.
- Other types (`docs:`, `test:`, `ci:`, `chore:` and so on) don't trigger a
  release by themselves.
- A commit counts towards a module if it changes files in that module.
  `vipercommand/`, `koanfcommand/` and `.github/` are excluded from the root
  module.
- Configuration lives in `release-please-config.json`. The current version of
  each module is in `.release-please-manifest.json`, which release-please
  updates itself.

When a `vipercommand` or `koanfcommand` change needs an unreleased change in
`simplecommand`, release in this order:

1. Merge the `simplecommand` change, then merge its release PR (creating
   `vX.Y.Z`).
2. In the module's directory, run
   `go get github.com/andrewheberle/simplecommand@vX.Y.Z && go mod tidy`.
3. Merge the module's change, then merge its release PR.

## Code conventions
- Go version: [1.26], per `go.mod`. Don't bump it unless asked.
- Dependencies: prefer the standard library. Ask before adding a new module dependency. Run `go mod tidy` after any dependency change.

## Running tests

Commands run from the root only cover the root module, so always run them in
every module:

```sh
go vet ./...
go test -race -cover ./...

cd vipercommand
go vet ./...
go test -race -cover ./...

cd ../koanfcommand
go vet ./...
go test -race -cover ./...
```

CI (`.github/workflows/ci.yml`) runs the same steps on pull requests and on
pushes to `main`. `go vet` and the tests must pass in every module before you
commit.

When editing workflows, pin every action to a full commit SHA with the
release as a comment (`uses: owner/action@<sha> # vX.Y.Z`). Renovate keeps
these up to date.

CI also runs the `vipercommand` and `koanfcommand` vet and tests against the
local `simplecommand` through a temporary `go.work`. This catches a
`simplecommand` change that would break either module before it is released.
If only this step fails, fix that module or rethink the `simplecommand`
change. Don't work around it with a `replace` directive. To reproduce it
locally (and likewise in `koanfcommand`):

```sh
cd vipercommand
GOWORK="$(mktemp -d)/go.work" sh -c \
  'go work init .. . && go vet ./... && go test -race ./...'
```

After changing dependencies, run `go mod tidy` in the affected module.

## Linting

CI runs [golangci-lint](https://golangci-lint.run/) v2 on every module. All
modules share the configuration in `.golangci.yml` at the repository root
(golangci-lint searches parent directories for it). Run it in every module
before committing:

```sh
golangci-lint run ./...

cd vipercommand
golangci-lint run ./...

cd ../koanfcommand
golangci-lint run ./...
```

Fix reported issues rather than suppressing them. If a `//nolint` directive is
unavoidable, name the linter and give a reason
(`//nolint:errcheck // reason`), which the `nolintlint` linter enforces.
Change `.golangci.yml` only when a rule is wrong for the whole codebase, and
add a comment there explaining why.

`golangci-lint fmt` applies the configured formatters (`gofmt` and
`goimports`).

## Code formatting

- Format all Go code with `gofmt` (or `go fmt ./...`). `gofmt -l .` must
  print nothing.
- Every exported identifier needs a doc comment that starts with its name.
  Use `[Name]` doc links to refer to other identifiers, as the existing code
  does.
- Keep the existing style: short lower-case `//` comments explaining intent
  above non-obvious blocks, and early returns for errors.
- Use UK English spelling in comments and documentation (for example
  "initialise"). The `misspell` linter is set to the UK locale.
- Never ignore an error that a function returns. Return it, or explain in a
  comment why it is safe to ignore.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>(<optional scope>): <description>

<optional body>

<optional footer(s)>
```

- **type**: one of `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `ci`,
  `build` or `chore`.
- **scope**: use `vipercommand` or `koanfcommand` for changes limited to
  that module, `deps` for dependency updates (as Renovate does), and omit it
  for the root module. There is no space between the type and the scope:
  `fix(vipercommand):`, not `fix (vipercommand):`.
- **description**: imperative mood, lower case, no trailing full stop, for
  example `fix(vipercommand): let command line flags take precedence`.
- **body**: explain what was wrong and why the change fixes it.
- **breaking changes**: add `!` after the type or scope (`feat!:`) and a
  `BREAKING CHANGE:` footer describing the migration.

Pull requests are squash-merged, so the PR title becomes the commit message
on `main` and must follow the same format. The `PR title` workflow
(`.github/workflows/pr-title.yml`) fails if it doesn't.

Keep each commit to one logical change. For example, put a bug fix and an
unrelated docs fix in separate commits.

**Keep changes to each module in separate pull requests.** A squash-merged
PR becomes one commit, and release-please counts that commit towards every
module whose files it changes, whatever its scope. Every file outside
`vipercommand/` and `koanfcommand/` belongs to the root module, apart from
`.github/` (excluded in `release-please-config.json`). So a
`fix(vipercommand):` or `feat(vipercommand):` PR must only change files under
`vipercommand/` or `.github/`, and likewise for `koanfcommand`. Otherwise it
also triggers a root module release. Put any other
changes, including to `AGENTS.md`, the README or `.golangci.yml`, in a
separate PR.

## Implementing tests

- **Examples are documentation.** `example_test.go` in each module holds
  `Example*` functions that appear on pkg.go.dev. They must have an
  `// Output:` comment so they run as tests. Keep them short, idiomatic and
  correct, because users copy them (for example, always check errors).
- **Behaviour tests go in `<package>_test.go`** (such as
  `vipercommand/vipercommand_test.go`). Use the external test package
  (`package vipercommand_test`) so tests use only the public API.
- **Prefer table-driven tests** with `t.Run` subtests, like
  `TestFlagPrecedence`.
- **Make tests deterministic.** Pass explicit arguments to `Execute`, never
  `os.Args`. Use `t.Setenv` for environment variables in tests. `Example`
  functions have no `*testing.T`, so they use `os.Setenv` with a deferred
  `os.Unsetenv` instead.
- **Use the existing fixtures**, such as `vipercommand/testconfig.yml` and
  `koanfcommand/testconfig.yml`,
  before adding new ones. Tests run with the package directory as the working
  directory, so use relative paths.
- **Every bug fix needs a regression test that fails without the fix.**
  Check this by running the test before you apply the fix.
- **For `vipercommand` and `koanfcommand`, test the precedence order:** command line flag, then
  environment variable, then configuration file, then flag default.
- Coverage is reported to Codecov. New code should not reduce coverage.

## Boundaries
- Don't commit secrets, tokens, or real config values.
- Don't modify generated files; regenerate them with `go generate ./...`.
- Keep changes scoped to the task. No drive-by refactors or renames.
