# Contributing to polygolem

Thanks for considering a contribution. This file describes how to build,
test, file an issue, and what we expect from a pull request.

## Build, test, lint

`polygolem` is a single Go module. Standard toolchain only.

```bash
# Build the binary
go build -o polygolem ./cmd/polygolem

# Run all tests
go test ./...

# Static analysis
go vet ./...

# Formatting (writes in place; CI fails if anything is reformatted)
gofmt -w .
```

The CI workflow at `.github/workflows/ci.yml` runs the same four steps on
every push and pull request.

## TDD-first discipline

`polygolem` is a TDD-first project. Behavior changes land with tests, and
new tests fail before the implementation lands. The test layout follows
the standard Go convention: `*_test.go` siblings inside each package, plus
end-to-end checks under `tests/`.

If a change cannot be expressed as a failing test first (rare — usually
docs-only or repo-meta), say so explicitly in the pull request.

## Documentation surfaces

`polygolem` keeps documentation in five places. When you change behavior,
update the surfaces that lose accuracy:

| Surface | Audience | Source |
|---|---|---|
| `README.md` | Drive-by readers, install + headline pitch. | Repo root. |
| `docs/*.md` | Operators and integrators (architecture, commands, safety, PRD). | `docs/`. |
| Astro docs site | Long-form web docs, search-indexed. | `docs-site/`. |
| `SKILL.md` | Agentic consumers (Claude Code skill manifest). | Repo root. |
| Godoc comments | Go SDK consumers (`pkg/`) and contributors (`internal/`). | Inline in `.go` files. |

A change to a CLI flag typically touches `README.md`, `docs/COMMANDS.md`,
the docs-site equivalent, and `SKILL.md`. A change to a `pkg/` API touches
godoc and the docs-site reference.

## Filing an issue

Open issues at https://github.com/TrebuchetDynamics/polygolem/issues.
Use the **Bug report** template for behavior bugs and the **Feature
request** template for proposals. Include the exact command you ran and
the JSON output (with `--json`) when applicable.

For security-sensitive reports — anything touching private keys, the
deposit-wallet flow, signing paths, or builder credentials — follow
`SECURITY.md` instead. Do not file public issues for those.

## Filing a pull request

Use the pull-request template. The checklist is short:

- Tests pass locally (`go test ./...`).
- Godoc updated if the exported surface changed (`pkg/` or exported
  identifiers in `internal/`).
- `CHANGELOG.md` `## [Unreleased]` section updated with a one-line
  description of the change.

Where the planning artifacts live, in case a PR references them:

- Specs: `docs/superpowers/specs/`
- Plans: `docs/superpowers/plans/`
- Working audit notes (when present): `docs/AUDIT-FINDINGS.md` (created
  per project, deleted when consumed).

Thanks for reading.
