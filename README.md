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
