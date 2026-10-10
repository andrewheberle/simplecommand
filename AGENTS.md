# AGENTS.md

Guidance for AI coding agents and contributors working in this repository.

## Repository layout

This repository contains **two separate Go modules**:

| Path | Module | Notes |
|------|--------|-------|
| `/` | `github.com/andrewheberle/simplecommand` | Core `Command` type. Minimal dependencies. Do not add `viper` here. |
| `vipercommand/` | `github.com/andrewheberle/simplecommand/vipercommand` | Viper-enabled `Command`. Has its own `go.mod` and `go.sum`. |

`vipercommand` depends on a **published** version of `simplecommand` (see
`vipercommand/go.mod`), not the local copy. Do not add a `replace` directive.
To test changes to both modules together, create a `go.work` file locally (it
is in `.gitignore` and must not be committed):

```sh
go work init . ./vipercommand
```

Releases are tagged per module: `vX.Y.Z` for the root module and
`vipercommand/vX.Y.Z` for the submodule.

## Running tests

Commands run from the root only cover the root module, so always run them in
both modules:

```sh
go vet ./...
go test -race -cover ./...

cd vipercommand
go vet ./...
go test -race -cover ./...
```

CI (`.github/workflows/codecov.yml`) runs the same steps on every push.
`go vet` and the tests must pass in both modules before you commit.

After changing dependencies, run `go mod tidy` in the affected module.

## Code formatting

- Format all Go code with `gofmt` (or `go fmt ./...`). `gofmt -l .` must
  print nothing.
- Every exported identifier needs a doc comment that starts with its name.
  Use `[Name]` doc links to refer to other identifiers, as the existing code
  does.
- Keep the existing style: short lower-case `//` comments explaining intent
  above non-obvious blocks, and early returns for errors.
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
- **scope**: use `vipercommand` for changes limited to that module, `deps`
  for dependency updates (as Renovate does), and omit it for the root module.
  There is no space between the type and the scope: `fix(vipercommand):`, not
  `fix (vipercommand):`.
- **description**: imperative mood, lower case, no trailing full stop, for
  example `fix(vipercommand): let command line flags take precedence`.
- **body**: explain what was wrong and why the change fixes it.
- **breaking changes**: add `!` after the type or scope (`feat!:`) and a
  `BREAKING CHANGE:` footer describing the migration.

Pull requests are squash-merged, so the PR title becomes the commit message
on `main` and must follow the same format.

Keep each commit to one logical change. For example, put a bug fix and an
unrelated docs fix in separate commits.

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
- **Use the existing fixtures**, such as `vipercommand/testconfig.yml`,
  before adding new ones. Tests run with the package directory as the working
  directory, so use relative paths.
- **Every bug fix needs a regression test that fails without the fix.**
  Check this by running the test before you apply the fix.
- **For `vipercommand`, test the precedence order:** command line flag, then
  environment variable, then configuration file, then flag default.
- Coverage is reported to Codecov. New code should not reduce coverage.
