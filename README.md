# School Canteen

**English** · [Português](README.pt-BR.md)

Backend for a prepaid school canteen. Guardians top up the student's wallet via Pix and set spending rules (blocked products or categories, daily limit); the canteen operator registers purchases, which are authorized against the balance and the rules in a single transaction. Balances are kept in an append-only, double-entry ledger.

No front end: the interface is the HTTP API (OpenAPI), a cashier CLI and simulators.

## Documentation

Detailed documentation is written in Portuguese.

- [Domain](docs/domain.md): actors, glossary, contexts, ledger, purchase flow, rules
- [Architecture](docs/architecture.md): containers and how they communicate
- [Use cases](docs/use-cases.md)
- [Non-functional requirements](docs/non-functional-requirements.md)
- [Architecture decision records (ADRs)](docs/adr/)
- [Conventions](docs/conventions.md): languages, Go, tests, data, process
- [Roadmap](docs/roadmap.md): phases, deliverables and done criteria
- [Board](https://github.com/users/betoth/projects/1): progress of the current phase

## Stack

| Tool | Why |
|---|---|
| Go 1.27 | current stable release, pinned in `go.mod` |
| PostgreSQL | transactions, `FOR UPDATE`, one schema per context |
| sqlc + pgx | explicit SQL with generated typed code ([ADR 0002](docs/adr/0002-acesso-a-dados.md)) |
| goose | versioned SQL migrations embedded in the binary ([ADR 0001](docs/adr/0001-migrations.md)) |
| Kafka (franz-go) | per-student ordering, independent consumers, replay ([ADR 0003](docs/adr/0003-mensageria.md)) |
| Outbox | event published only if the transaction commits ([ADR 0004](docs/adr/0004-outbox.md)) |
| OpenTelemetry + Collector | vendor-neutral instrumentation; backends (Jaeger, Prometheus) swapped by Collector config |
| golangci-lint | standard Go linter suite; official binary |
| `tools/go.mod` | versioned Go tools without mixing their dependencies into the application's |

## AI cost

Claude Code token cost is measured per issue and stored in [`docs/ai-costs.csv`](docs/ai-costs.csv), as numbers only (no conversation content). The issue comes from the branch name (`feat/42-...` counts toward #42).

To record sessions, once per machine:

1. `make tokencost`, which installs the CLI used by the Claude Code hooks in `.claude/settings.json`. The hooks skip silently when the CLI is not found, so `$(go env GOPATH)/bin` must be in the `PATH`.
2. Raise `cleanupPeriodDays` in `~/.claude/settings.json` (default is 30 days), so transcripts are not deleted before collection.

`make costs` updates the CSV and the charts in [`docs/ai-costs.md`](docs/ai-costs.md). Sessions from before the hooks are imported with `go -C tools run ./tokencost collect -root "$PWD" -import-old`.
