## Summary

One or two sentences describing what this PR does and why.

## Changes

- Bullet list of the substantive changes.

## Checklist

- [ ] Tests pass locally (`go test ./...`).
- [ ] `go vet ./...` is clean.
- [ ] `git ls-files -z '*.go' ':!:opensource-projects/**' | xargs -0 gofmt -l`
      produces no output.
- [ ] Godoc updated if the exported surface changed (`pkg/` or exported
      identifiers in `internal/`).
- [ ] `CHANGELOG.md` `## [Unreleased]` section updated with a one-line
      entry.
- [ ] Doc surfaces (`README.md`, `docs/*.md`, `docs-site/`, `SKILL.md`)
      updated where this change affects them.
- [ ] No secrets, private keys, builder credentials, or `.env` files
      committed.

## Test plan

How a reviewer can verify this change locally. Commands, expected
output, sample data.

## Related

Issue numbers, prior PRs, or specs/plans under `docs/superpowers/`.
