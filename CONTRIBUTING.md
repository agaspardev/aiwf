# Contributing to aiwf

Thanks for your interest! aiwf is a CLI tool to install and manage AI coding agents with Spec-Driven Development (SDD), persistent memory, and security pipelines.

## Quick start

```bash
git clone https://github.com/agaspardev/aiwf.git
cd aiwf
go build ./cmd/aiwf
```

## Requirements

- Go 1.25+
- `gopkg.in/yaml.v3` (single external dependency)

## Before submitting

1. Run `go test ./...` — all tests must pass
2. Run `go build ./...` — no compilation errors
3. Run `go vet ./...` — no warnings
4. Keep `THIRD_PARTY_NOTICES.md` updated if adding dependencies

## Commit conventions

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(cli): add new command
fix(security): correct severity classification
docs(readme): update usage examples
refactor(overlay): simplify merge logic
```

## Code review

All changes go through the project's SDD workflow:

1. Proposal → Spec/Design → Tasks → Apply → Verify → Archive
2. The quality gate (`/quality-gate`) must pass before merge
3. Changes without SDD context may be requested to follow the process

## License

MIT — see [LICENSE](LICENSE)
